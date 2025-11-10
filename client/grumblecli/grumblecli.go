package grumblecli

import (
	"fmt"
	"io"

	"github.com/desertbit/grumble"
	"github.com/olekukonko/tablewriter"
	"google.golang.org/grpc"

	"orsted/protobuf/orstedrpc"
)

var app = grumble.New(&grumble.Config{
	Name:        "orsted-client",
	Description: "Client to orsted app",
	HistoryFile: "orsted.history",
})

var SelectedSession *orstedrpc.Session = nil

func prettyPrint(data [][]string, headers []string, out io.Writer) {
	table := tablewriter.NewWriter(out)
	table.SetHeader(headers)
	for _, v := range data {
		table.Append(v)
	}
	table.Render()
}

func addSingleCommandFromString(commandString string, conn grpc.ClientConnInterface) {
	switch commandString {
	case "generate":
		SetGenerateBeaconCommand()
	case "listener":
		SetListenerCommands(conn)
	case "session":
		SetSessionCommands(conn)
	case "task":
		SetGeneralTaskCommands(conn)
	case "load-module":
		SetLoadModuleCommand(conn)
	case "run":
		SetRunCommand(conn)
	case "ps":
		SetPsCommands(conn)
	case "interact":
		SetInteractCommands(conn)
	case "inline-clr":
		SetInlineExecCommand(conn)
	case "pivot":
		SetPivotCommand(conn)
	case "socks":
		SetSocksCommand(conn)
	case "download":
		SetDownloadCommands(conn)
	case "upload":
		SetUploadCommands(conn)
	case "evasion":
		SetEvasionCommand(conn)
	case "execute-assembly":
		SetAssembluExecCommand(conn)
	case "shell":
		SetShellCommands(conn)
	case "sleep":
		SetSleepCommand(conn)
	case "back":
		SetBackCommands(conn)
	case "hoster":
		SetHosterCommands(conn)
	case "token":
		SetTokenCommands(conn)
	case "runas":
		SetRunAsCommand(conn)
	case "ls":
		SetlsCommand(conn)
	case "autoroute":
		SetAutorouteCommand(conn)
	case "procdump":
		SetProcDumpCommand(conn)
	case "cat":
		SetCatCommand(conn)
	case "psexec":
		SetPsExecCommand(conn)
	case "powercliff":
		SetPowercliffCommand(conn)
	case "winrm":
		SetWinrmCommand(conn)
	case "help":
		// Use default grumble help
		return
	default:
		fmt.Println("Command not found ", commandString)

	}
}

func SetCommands(conn grpc.ClientConnInterface) {
	// Removing All Commands
	for _, v := range Conf.LinuxCommands {
		app.Commands().Remove(v)
	}
	for _, v := range Conf.GlobalCommands {
		app.Commands().Remove(v)
	}
	for _, v := range Conf.WindowsCommands {
		app.Commands().Remove(v)
	}

	// Adding all commands depending on context
	var listOfCommand []string
	if SelectedSession == nil {
		listOfCommand = Conf.GlobalCommands
		for _, v := range listOfCommand {
			addSingleCommandFromString(v, conn)
		}
		return

	}

	if SelectedSession.Os == "linux" {
		listOfCommand = Conf.LinuxCommands
	}
	if SelectedSession.Os == "windows" {
		listOfCommand = Conf.WindowsCommands
	}
	for _, v := range listOfCommand {
		addSingleCommandFromString(v, conn)
	}

	return
}

func Run() {
	app.Run()
}
