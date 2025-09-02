package clientrpc

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"

	"orsted/protobuf/orstedrpc"
)

func StartHttpListenerFunc(conn grpc.ClientConnInterface, ip string, port string) (*orstedrpc.ListenerJob, error) {
	c := orstedrpc.NewOrstedRpcClient(conn)
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
    res, err := c.StartListener(ctx, &orstedrpc.ListenerReq{ListenerType: "http", Ip: ip, Port: port})
	return res, err
}

func StartHttpsListenerFunc(conn grpc.ClientConnInterface, ip string, port string, certpath string, keypath string) (*orstedrpc.ListenerJob, error) {
	c := orstedrpc.NewOrstedRpcClient(conn)
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
    res, err := c.StartListener(ctx, &orstedrpc.ListenerReq{ListenerType: "https", Ip: ip, Port: port, CertPath: certpath, KeyPath: keypath})
	return res, err
}

func ListListenerFunc(conn grpc.ClientConnInterface) (*orstedrpc.ListenerJobList, error) {
	c := orstedrpc.NewOrstedRpcClient(conn)
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	res, err := c.ListListeners(ctx, &orstedrpc.EmptyMessage{})
	return res, err
}

func DeleteListenerFunc(conn grpc.ClientConnInterface, id string) error {
	c := orstedrpc.NewOrstedRpcClient(conn)
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	_, err := c.DeleteListener(ctx, &orstedrpc.ListenerJob{Id: id, Ip: "", Port: ""})
	return err
}

func ListSessionFunc(conn grpc.ClientConnInterface) (*orstedrpc.SessionList, error) {
	c := orstedrpc.NewOrstedRpcClient(conn)
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	res, err := c.ListSessions(ctx, &orstedrpc.EmptyMessage{})
	return res, err
}

func GetSessionByIdFunc(conn grpc.ClientConnInterface, id string) (*orstedrpc.Session, error) {
	c := orstedrpc.NewOrstedRpcClient(conn)
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	res, err := c.GetSessionById(ctx, &orstedrpc.IdMessage{Id: id})
	return res, err
}

func AddTaskFunc(conn grpc.ClientConnInterface, sessionId string, command string, reqdata []byte, prettyCommand string) (*orstedrpc.Task, error) {
	c := orstedrpc.NewOrstedRpcClient(conn)
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	res, err := c.AddTask(ctx, &orstedrpc.TaskReq{BeacondId: sessionId, Command: command, Reqdata: reqdata, PrettyCommand: prettyCommand})
	return res, err
}

func ListTaskFunc(conn grpc.ClientConnInterface, beaconId string) (*orstedrpc.TaskList, error){
	c := orstedrpc.NewOrstedRpcClient(conn)
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	res, err := c.ListTask(ctx, &orstedrpc.Session{Id: beaconId})
	return res, err
}

func GetSingleTaskFunc(conn grpc.ClientConnInterface, taskId string) (*orstedrpc.Task, error) {
	c := orstedrpc.NewOrstedRpcClient(conn)
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
    res, err := c.GetSingleTask(ctx, &orstedrpc.Task{TaskId: taskId})
    return res, err
}


func StartSocks(conn grpc.ClientConnInterface, beaconId string, ip string, port string) (*orstedrpc.Socks, error) {
	c := orstedrpc.NewOrstedRpcClient(conn)
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
    res, err := c.StartSocks(ctx, &orstedrpc.Socks{BeaconId: beaconId, Ip: ip, Port: port, Status: "pending"})
    return res, err
}

func HostFileFunc(conn grpc.ClientConnInterface, filename string, data []byte) (error) {
	c := orstedrpc.NewOrstedRpcClient(conn)
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
    res, err := c.HostFile(ctx, &orstedrpc.Host{Filename: filename, Data: data})
	if err != nil {
		return err
	}
	if res.Result == -1 {
		return fmt.Errorf("Some Error occured when hosting file")
	}
    return nil
}

func UnHostFileFunc(conn grpc.ClientConnInterface, filename string) (error) {
	c := orstedrpc.NewOrstedRpcClient(conn)
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
    res, err := c.UnHostFile(ctx, &orstedrpc.Host{Filename: filename})
	if err != nil {
		return err
	}
	if res.Result == -1 {
		return fmt.Errorf("Some Error occured when unhosting file")
	}
    return nil
}

func ViewHostFileFunc(conn grpc.ClientConnInterface) (*orstedrpc.HostList, error) {
	c := orstedrpc.NewOrstedRpcClient(conn)
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
    res, err := c.ViewHostFile(ctx, &orstedrpc.EmptyMessage{})
    return res, err
}

func AddRoute(conn grpc.ClientConnInterface, beaconId string, subnet string) (*orstedrpc.Route, error) {
	c := orstedrpc.NewOrstedRpcClient(conn)
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
    res, err := c.AddRoute(ctx, &orstedrpc.RouteReq{BeaconId: beaconId, Subnet: subnet})
    return res, err
}

func ListRoute(conn grpc.ClientConnInterface) (*orstedrpc.RouteList, error) {
	c := orstedrpc.NewOrstedRpcClient(conn)
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
    res, err := c.ListRoute(ctx, &orstedrpc.EmptyMessage{})
    return res, err
}

func DeleteRoute(conn grpc.ClientConnInterface, beaconId string, subnet string) (*orstedrpc.ResultMessage, error) {
	c := orstedrpc.NewOrstedRpcClient(conn)
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
    res, err := c.DeleteRoute(ctx, &orstedrpc.Route{BeaconId: beaconId, Subnet: subnet})
    return res, err
}
