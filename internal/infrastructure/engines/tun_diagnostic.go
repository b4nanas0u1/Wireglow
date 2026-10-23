package engines

// This fixed capability error is safe to publish in a controller snapshot.
// Generic preparation errors may contain private profile data and stay hidden.
type windowsTURNUDPError struct{}

func (windowsTURNUDPError) Error() string {
	return "tun: Windows TURN UDP socket preservation is not supported yet; select TCP TURN transport"
}
func (e windowsTURNUDPError) PublicDiagnostic() string { return e.Error() }
