package pivot

import "fmt"

type Pivot interface {
	Start() error
	Stop()
	GetID() string
	GetType() string
	GetString() string
}

var PIVOT_LIST []Pivot
var Pivox int = 0

func StartPivot(pivType string, address string) {
	switch pivType {
	case "tcp":
		t, err := NewTcpStuff(address)
		if err != nil {
			return
		}
		PIVOT_LIST = append(PIVOT_LIST, &t)
		err = t.Start()
		if err != nil {
			for i, piv := range PIVOT_LIST {
				if piv.GetID() == t.Id {
					piv.Stop()
					PIVOT_LIST = append(PIVOT_LIST[:i], PIVOT_LIST[i+1:]...)
					return
				}
			}
		}

	}
}

func StopPivot(pivID string) error {
	for i, piv := range PIVOT_LIST {
		if piv.GetID() == pivID {
			piv.Stop()
			PIVOT_LIST = append(PIVOT_LIST[:i], PIVOT_LIST[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("Pivot ID could not be found")
}

func ListPivot() string {
	res := "List of pivot:\n"
	res += "--------------\n"
	res += "--------------\n"
	for _, piv := range PIVOT_LIST {
		pivStr := fmt.Sprintf("Pivot %s: %s (%s)\n", piv.GetID(), piv.GetString(), piv.GetType())
		res += pivStr
	}
	return res
}
