package entity

type MobilityProfile string

const (
	MobilityWheelchair MobilityProfile = "wheelchair"
	MobilityStroller   MobilityProfile = "stroller"
	MobilityElderly    MobilityProfile = "elderly"
	MobilityDefault    MobilityProfile = "default"
)

func (m MobilityProfile) IsValid() bool {
	switch m {
	case MobilityWheelchair, MobilityStroller, MobilityElderly, MobilityDefault:
		return true
	default:
		return false
	}
}

func (m MobilityProfile) MaxSeverity() int {
	switch m {
	case MobilityWheelchair:
		return 2
	case MobilityStroller:
		return 3
	case MobilityElderly:
		return 4
	case MobilityDefault:
		return 5
	default:
		return 5
	}
}