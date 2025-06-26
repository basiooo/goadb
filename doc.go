/*
package adb is a Go interface to the Android Debug Bridge (adb).

Example usage:

	client, err := adb.New()
	if err != nil {
		log.Fatal(err)
	}

	// List all connected devices
	devices, err := client.ListDevices()
	if err != nil {
		log.Fatal(err)
	}

	// Connect to a specific device
	device := client.Device(adb.DeviceWithSerial("device_serial"))

	// Run a command on the device
	output, err := device.RunCommand("ls", "-la")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(output)

The client/server spec is defined at https://android.googlesource.com/platform/system/core/+/master/adb/OVERVIEW.TXT.

WARNING This library is under heavy development, and its API is likely to change without notice.
*/
package adb

// TODO(z): Write method-specific examples.
