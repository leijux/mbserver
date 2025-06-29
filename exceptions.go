package mbserver

//go:generate stringer -type=Exception

// Exception codes.
type Exception uint8

var (
	// Success operation successful.
	Success Exception = 0

	// IllegalFunction function code received in the query is not recognized or allowed by slave.
	IllegalFunction Exception = 1

	// IllegalDataAddress data address of some or all the required entities are not allowed or do not exist in slave.
	IllegalDataAddress Exception = 2

	// IllegalDataValue value is not accepted by slave.
	IllegalDataValue Exception = 3

	// SlaveDeviceFailure Unrecoverable error occurred while slave was attempting to perform requested action.
	SlaveDeviceFailure Exception = 4

	// AcknowledgeSlave has accepted request and is processing it, but a long duration of time is required. This response is returned to prevent a timeout error from occurring in the master. Master can next issue a Poll Program Complete message to determine whether processing is completed.
	AcknowledgeSlave Exception = 5

	// SlaveDeviceBusy is engaged in processing a long-duration command. Master should retry later.
	SlaveDeviceBusy Exception = 6

	// NegativeAcknowledge Slave cannot perform the programming functions. Master should request diagnostic or error information from slave.
	NegativeAcknowledge Exception = 7

	// MemoryParityError Slave detected a parity error in memory. Master can retry the request, but service may be required on the slave device.
	MemoryParityError Exception = 8

	// GatewayPathUnavailable Specialized for Modbus gateways. Indicates a misconfigured gateway.
	GatewayPathUnavailable Exception = 10

	// GatewayTargetDeviceFailedToRespond Specialized for Modbus gateways. Sent when slave fails to respond.
	GatewayTargetDeviceFailedToRespond Exception = 11
)

func (e Exception) Error() string {
	return e.String()
}
