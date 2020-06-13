package authapi

func (*SMSInfo) isSecondFactorInfo()        {}
func (*TOTPInfo) isSecondFactorInfo()       {}
func (*BackupCodeInfo) isSecondFactorInfo() {}

// SecondFactorInfo is an interface of second factors for the application.
type SecondFactorInfo interface {
	isSecondFactorInfo()
}
