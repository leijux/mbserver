package mbserver

//go:generate stringer -type=Exception

// Exception codes.
type Exception uint8

const (
	// Success operation successful.
	Success Exception = iota

	// IllegalFunction function code received in the query is not recognized or allowed by slave.
	IllegalFunction

	// IllegalDataAddress data address of some or all the required entities are not allowed or do not exist in slave.
	IllegalDataAddress

	// IllegalDataValue value is not accepted by slave.
	IllegalDataValue

	// SlaveDeviceFailure Unrecoverable error occurred while slave was attempting to perform requested action.
	SlaveDeviceFailure

	// AcknowledgeSlave has accepted request and is processing it, but a long duration of time is required. This response is returned to prevent a timeout error from occurring in the master. Master can next issue a Poll Program Complete message to determine whether processing is completed.
	AcknowledgeSlave

	// SlaveDeviceBusy is engaged in processing a long-duration command. Master should retry later.
	SlaveDeviceBusy

	// NegativeAcknowledge Slave cannot perform the programming functions. Master should request diagnostic or error information from slave.
	NegativeAcknowledge

	// MemoryParityError Slave detected a parity error in memory. Master can retry the request, but service may be required on the slave device.
	MemoryParityError

	// GatewayPathUnavailable Specialized for Modbus gateways. Indicates a misconfigured gateway.
	GatewayPathUnavailable Exception = 10

	// GatewayTargetDeviceFailedToRespond Specialized for Modbus gateways. Sent when slave fails to respond.
	GatewayTargetDeviceFailedToRespond Exception = 11
)

func (e Exception) Error() string {
	return e.String()
}
