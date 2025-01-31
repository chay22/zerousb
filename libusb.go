// go:build (freebsd && cgo) || (linux && cgo) || (darwin && !ios && cgo) || (windows && cgo)

package zerousb

/*
	#include "./libusb/libusb/libusb.h"
	// ctx is a global libusb context to interact with devices through.
	libusb_context* ctx;
*/
import "C"

import (
	"fmt"
	"reflect"
	"sync"
	"unsafe"
)

type libusbContext C.libusb_context

type Context struct {
	ctx    *libusbContext
	done   chan struct{}
	libusb libusbDevice

	mu      sync.Mutex
	devices map[*Device]bool
}

// libusbDevice is a USB connected device handle.
type libusbDevice struct {
	DeviceInfo // Embed the infos for easier access

	handle       *C.struct_libusb_device_handle // Low level USB device to communicate through
	lock         sync.Mutex
	writeTimeout int
	readTimeout  int
}

// enumerateRawWithRef is the internal device enumerator that retains 1 reference
// to every matched device so they may selectively be opened on request.
func getAllDevices(vendorID ID, productID ID) ([]DeviceInfo, error) {
	// Ensure we have a libusb context to interact through. The enumerate call is
	// protected by a mutex outside, so it's fine to do the below check and init.
	if C.ctx == nil {
		if err := fromLibusbErrno(C.libusb_init((**C.libusb_context)(&C.ctx))); err != nil {
			return nil, fmt.Errorf("failed to initialize libusb: %v", err)
		}
	}

	// Retrieve all the available USB devices and wrap them in Go
	var deviceList **C.libusb_device
	defer C.libusb_free_device_list(deviceList, 1)

	count := C.libusb_get_device_list(C.ctx, &deviceList)

	if count < 0 {
		return nil, libusbError(count)
	}

	var devices []*C.libusb_device
	*(*reflect.SliceHeader)(unsafe.Pointer(&devices)) = reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(deviceList)),
		Len:  int(count),
		Cap:  int(count),
	}

	var infos []DeviceInfo
	for devnum, dev := range devices {
		// Retrieve the libusb device descriptor and skip non-queried ones
		var desc C.struct_libusb_device_descriptor
		if err := fromLibusbErrno(C.libusb_get_device_descriptor(dev, &desc)); err != nil {
			return infos, fmt.Errorf("failed to get device %d descriptor: %v", devnum, err)
		}
		if (vendorID > 0 && ID(desc.idVendor) != vendorID) || (productID > 0 && ID(desc.idProduct) != productID) {
			continue
		}
		// Skip HID devices, they are handled directly by OS libraries
		if desc.bDeviceClass == C.LIBUSB_CLASS_HID {
			continue
		}
		// Iterate over all the configurations and find raw interfaces
		for cfgnum := 0; cfgnum < int(desc.bNumConfigurations); cfgnum++ {
			// Retrieve the all the possible USB configurations of the device
			var cfg *C.struct_libusb_config_descriptor
			if err := fromLibusbErrno(C.libusb_get_config_descriptor(dev, C.uint8_t(cfgnum), &cfg)); err != nil {
				return infos, fmt.Errorf("failed to get device %d config %d: %v", devnum, cfgnum, err)
			}
			var ifaces []C.struct_libusb_interface
			*(*reflect.SliceHeader)(unsafe.Pointer(&ifaces)) = reflect.SliceHeader{
				Data: uintptr(unsafe.Pointer(cfg._interface)),
				Len:  int(cfg.bNumInterfaces),
				Cap:  int(cfg.bNumInterfaces),
			}
			// Drill down into each advertised interface
			for ifacenum, iface := range ifaces {
				if iface.num_altsetting == 0 {
					continue
				}
				var alts []C.struct_libusb_interface_descriptor
				*(*reflect.SliceHeader)(unsafe.Pointer(&alts)) = reflect.SliceHeader{
					Data: uintptr(unsafe.Pointer(iface.altsetting)),
					Len:  int(iface.num_altsetting),
					Cap:  int(iface.num_altsetting),
				}
				for _, alt := range alts {
					// Skip HID interfaces, they are handled directly by OS libraries
					if alt.bInterfaceClass == C.LIBUSB_CLASS_HID {
						continue
					}
					// Find the endpoints that can speak libusb interrupts
					var ends []C.struct_libusb_endpoint_descriptor
					*(*reflect.SliceHeader)(unsafe.Pointer(&ends)) = reflect.SliceHeader{
						Data: uintptr(unsafe.Pointer(alt.endpoint)),
						Len:  int(alt.bNumEndpoints),
						Cap:  int(alt.bNumEndpoints),
					}
					var reader, writer *uint8
					var readerTransferType, writerTransferType uint8
					for _, end := range ends {
						// Skip any non-interrupt and bulk endpoints
						if end.bmAttributes != C.LIBUSB_TRANSFER_TYPE_INTERRUPT && end.bmAttributes != C.LIBUSB_TRANSFER_TYPE_BULK {
							continue
						}
						if end.bEndpointAddress&C.LIBUSB_ENDPOINT_IN == C.LIBUSB_ENDPOINT_IN {
							reader = new(uint8)
							*reader = uint8(end.bEndpointAddress)
							readerTransferType = uint8(end.bmAttributes)
						} else {
							writer = new(uint8)
							*writer = uint8(end.bEndpointAddress)
							writerTransferType = uint8(end.bmAttributes)
						}
					}
					// If both in and out interrupts are available, match the device
					if reader != nil && writer != nil {
						// Enumeration matched, bump the device refcount to avoid cleaning it up
						C.libusb_ref_device(dev)

						port := uint8(C.libusb_get_port_number(dev))
						info := DeviceInfo{
							Path:               fmt.Sprintf("%04x:%04x:%02d", vendorID, uint16(desc.idProduct), port),
							VendorID:           uint16(desc.idVendor),
							ProductID:          uint16(desc.idProduct),
							Class:              uint8(desc.bDeviceClass),
							SubClass:           uint8(desc.bDeviceSubClass),
							Protocol:           uint8(desc.bDeviceProtocol),
							Interface:          ifacenum,
							libusbDevice:       dev,
							libusbPort:         &port,
							libusbReader:       reader,
							libusbWriter:       writer,
							readerTransferType: &readerTransferType,
							writerTransferType: &writerTransferType,

							InterfaceAlternate: int(alt.bAlternateSetting),
							InterfaceClass:     uint8(alt.bInterfaceClass),
							InterfaceSubClass:  uint8(alt.bInterfaceSubClass),
							InterfaceProtocol:  uint8(alt.bInterfaceProtocol),
						}
						infos = append(infos, info)
					}
				}
			}
		}
	}

	for _, info := range infos {
		C.libusb_unref_device(info.libusbDevice.(*C.libusb_device))
	}

	return infos, nil
}

// open connects to a libusb device by its path name.
func open(info DeviceInfo) (*libusbDevice, error) {
	matches, err := getAllDevices(ID(info.VendorID), ID(info.ProductID))
	if err != nil {
		for _, match := range matches {
			C.libusb_unref_device(match.libusbDevice.(*C.libusb_device))
		}
		return nil, err
	}

	var device *C.libusb_device
	for _, match := range matches {
		// Keep the matching device reference, release anything else
		if device == nil && *match.libusbPort == *info.libusbPort && match.Interface == info.Interface {
			device = match.libusbDevice.(*C.libusb_device)
		} else {
			C.libusb_unref_device(match.libusbDevice.(*C.libusb_device))
		}
	}

	if device == nil {
		return nil, fmt.Errorf("failed to open device: not found")
	}

	info.libusbDevice = device

	var handle *C.struct_libusb_device_handle
	if err := fromLibusbErrno(C.libusb_open(info.libusbDevice.(*C.libusb_device), (**C.struct_libusb_device_handle)(&handle))); err != nil {
		return nil, fmt.Errorf("failed to open device: %v", err)
	}

	libusbDvc := &libusbDevice{
		DeviceInfo: info,
		handle:     handle,
	}

	libusbDvc.SetAutoDetach(1)
	libusbDvc.DetachKernelDriver()

	if err := fromLibusbErrno(C.libusb_claim_interface(handle, (C.int)(info.Interface))); err != nil {
		C.libusb_close(handle)
		return nil, fmt.Errorf("failed to claim interface: %v", err)
	}

	return &libusbDevice{
		DeviceInfo: info,
		handle:     handle,
	}, nil
}

// Close releases the raw USB device handle.
func (dev *libusbDevice) Close() error {
	dev.lock.Lock()
	defer dev.lock.Unlock()

	if dev.handle != nil {
		C.libusb_release_interface(dev.handle, (C.int)(dev.Interface))
		C.libusb_close(dev.handle)
		dev.handle = nil
	}
	C.libusb_unref_device(dev.libusbDevice.(*C.libusb_device))

	return nil
}

func (dev *libusbDevice) SetWriteTimeout(timeout int) {
	dev.writeTimeout = timeout
}

func (dev *libusbDevice) SetReadTimeout(timeout int) {
	dev.readTimeout = timeout
}

// Write sends a binary blob to an USB device.
func (dev *libusbDevice) Write(b []byte) (int, error) {
	dev.lock.Lock()
	defer dev.lock.Unlock()

	timeout := dev.writeTimeout

	switch *dev.writerTransferType {
	case C.LIBUSB_TRANSFER_TYPE_INTERRUPT:
		return dev.writeInterrupt(b, timeout)
	case C.LIBUSB_TRANSFER_TYPE_BULK:
		return dev.writeBulk(b, timeout)
	}

	return 0, fmt.Errorf("device transfer type unsupported %v", dev.readerTransferType)
}

// Read retrieves a binary blob from an USB device.
func (dev *libusbDevice) Read(b []byte) (int, error) {
	dev.lock.Lock()
	defer dev.lock.Unlock()

	timeout := dev.readTimeout

	switch *dev.readerTransferType {
	case C.LIBUSB_TRANSFER_TYPE_INTERRUPT:
		return dev.readInterrupt(b, timeout)
	case C.LIBUSB_TRANSFER_TYPE_BULK:
		return dev.readBulk(b, timeout)
	}

	return 0, fmt.Errorf("device transfer type unsupported %v", dev.readerTransferType)
}

func (dev *libusbDevice) SetAutoDetach(val int) error {
	err := fromLibusbErrno(C.libusb_set_auto_detach_kernel_driver(dev.handle, C.int(val)))
	if err != nil && err != ErrNotSupported {
		return err
	}
	return nil
}

func (dev *libusbDevice) DetachKernelDriver() error {
	err := fromLibusbErrno(C.libusb_detach_kernel_driver(dev.handle, C.int(dev.Interface)))
	if err != nil && err != ErrNotSupported && err != ErrNotFound {
		// ErrorNotSupported is returned in non linux systems
		// ErrorNotFound is returned if libusb's driver is already attached to the device
		return err
	}
	return nil
}

func (dev *libusbDevice) readInterrupt(b []byte, timeout int) (int, error) {
	var transferred C.int
	if err := fromLibusbErrno(C.libusb_interrupt_transfer(dev.handle, (C.uchar)(*dev.libusbReader), (*C.uchar)(&b[0]), (C.int)(len(b)), &transferred, (C.uint)(timeout))); err != nil {
		return 0, fmt.Errorf("failed to read from device: %v", err)
	}
	return int(transferred), nil
}

func (dev *libusbDevice) readBulk(b []byte, timeout int) (int, error) {
	var transferred C.int
	if err := fromLibusbErrno(C.libusb_bulk_transfer(dev.handle, (C.uchar)(*dev.libusbReader), (*C.uchar)(&b[0]), (C.int)(len(b)), &transferred, (C.uint)(timeout))); err != nil {
		return 0, fmt.Errorf("failed to read from device: %v", err)
	}
	return int(transferred), nil
}

func (dev *libusbDevice) writeBulk(b []byte, timeout int) (int, error) {
	var transferred C.int
	if err := fromLibusbErrno(C.libusb_bulk_transfer(dev.handle, (C.uchar)(*dev.libusbWriter), (*C.uchar)(&b[0]), (C.int)(len(b)), &transferred, (C.uint)(timeout))); err != nil {
		return 0, fmt.Errorf("failed to write to device: %v", err)
	}
	return int(transferred), nil
}

func (dev *libusbDevice) writeInterrupt(b []byte, timeout int) (int, error) {
	var transferred C.int
	if err := fromLibusbErrno(C.libusb_interrupt_transfer(dev.handle, (C.uchar)(*dev.libusbWriter), (*C.uchar)(&b[0]), (C.int)(len(b)), &transferred, (C.uint)(timeout))); err != nil {
		return 0, fmt.Errorf("failed to write to device: %v", err)
	}
	return int(transferred), nil
}
