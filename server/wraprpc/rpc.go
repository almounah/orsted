package wraprpc

import (
	"context"
	"errors"
	"fmt"

	"orsted/profiles"
	"orsted/protobuf/orstedrpc"
	"orsted/server/autoroute"
	"orsted/server/listeners"
	"orsted/server/orsteddb"
	"orsted/server/socks"
	"orsted/server/utils"
)

type Server struct {
	orstedrpc.UnimplementedOrstedRpcServer
}

func (s *Server) StartListener(c context.Context, req *orstedrpc.ListenerReq) (*orstedrpc.ListenerJob, error) {
	utils.PrintDebug("Got a RPC Req")
	utils.PrintDebug(req.Ip, req.Port, req.ListenerType)

    if isTcpAddressInUse(req.Ip + ":" + req.Port) {
        return nil, errors.New("Address Already in Use")
    }
	res, err := orsteddb.AddListener(req)
	if err != nil {
		utils.PrintDebug(err)
	}
    var l listeners.Listeners
    switch req.ListenerType {
    case "http":
        l = listeners.NewHttpListener(res.Id, req.Ip, req.Port)
    case "https":
        l = listeners.NewHttpsListener(res.Id, req.Ip, req.Port, req.CertPath, req.KeyPath)
        
    }
    go l.StartListener()
    listeners.LISTENERS_LIST = append(listeners.LISTENERS_LIST, l)

	return res, nil
}

func (s *Server) ListListeners(c context.Context, m *orstedrpc.EmptyMessage) (*orstedrpc.ListenerJobList, error) {
	res, err := orsteddb.ListListener()
	if err != nil {
		utils.PrintDebug("Error list listener from DB", err.Error())
	}
	return res, nil
}

func (s *Server) DeleteListener(c context.Context, m *orstedrpc.ListenerJob) (*orstedrpc.EmptyMessage, error) {
    err := listeners.DeleteListenerById(m.Id)
    if err != nil {
		utils.PrintDebug("Error Delete Listener from process", err.Error())
        return nil, err
    }
    err = orsteddb.DeleteListener(m)
    if err != nil {
		utils.PrintDebug("Error Delete Listener from DB", err.Error())
        return nil, err
    }
    return nil, err
}

func (s *Server) ListSessions(c context.Context, m *orstedrpc.EmptyMessage) (*orstedrpc.SessionList, error) {
	res, err := orsteddb.ListSessionsDb()
	if err != nil {
		utils.PrintDebug("Error List Session from DB", err.Error())
	}
	return res, nil
}

func (s *Server) GetSessionById(c context.Context, i *orstedrpc.IdMessage) (*orstedrpc.Session, error) {
	res, err := orsteddb.GetSessionById(i.Id)
	if err != nil {
		utils.PrintDebug("Error Get Session Id from DB", err.Error())
	}
	return res, nil
}

func (s *Server) StopSessionById(c context.Context, i *orstedrpc.IdMessage) (*orstedrpc.EmptyMessage, error) {
	err := orsteddb.ChangeBeaconStatus(i.Id, "stopped")
	if err != nil {
		utils.PrintDebug("Error Get Session Id from DB", err.Error())
	}

	var t orstedrpc.TaskReq
	t.BeacondId = i.Id
	t.Command = "stop"
	t.PrettyCommand = "Automatic Command: stop"

	s.AddTask(c, &t)
	return nil, nil
}

func (s *Server) AddTask(c context.Context, treq *orstedrpc.TaskReq) (*orstedrpc.Task, error) {
    res, err := orsteddb.AddTaskDb(treq)
	if err != nil {
		utils.PrintDebug("Error Add Task to DB", err.Error())
	}
	return res, nil
}

func (s *Server) ListTask(c context.Context, session *orstedrpc.Session) (*orstedrpc.TaskList, error) {
    res, err := orsteddb.ListTasksDb(session.Id, []string{"sent", "pending", "completed", "failed"})
	if err != nil {
		utils.PrintDebug("Error List Task from DB", err.Error())
	}
	return res, nil
}

func (s *Server) GetSingleTask(c context.Context, task *orstedrpc.Task) (*orstedrpc.Task, error) {
    res, err := orsteddb.GetTaskByIdDb(task.TaskId)
	if err != nil {
		utils.PrintDebug("Error Get Single Task from DB", err.Error())
	}
	return res, nil
}

func (s *Server) StartSocks(ctx context.Context, soc *orstedrpc.Socks) (*orstedrpc.Socks, error) {
    soc.Status = "running"
    go socks.StartNewSocksServer(soc.BeaconId, soc.Ip, soc.Port)
    return soc, nil
}

func (s* Server) HostFile(c context.Context, h *orstedrpc.Host) (*orstedrpc.ResultMessage, error) {
	err := orsteddb.HostFileDb(h)
	return &orstedrpc.ResultMessage{Result: 0}, err

}

func (s* Server) UnHostFile(c context.Context, h *orstedrpc.Host) (*orstedrpc.ResultMessage, error) {
	err := orsteddb.UnHostFileDb(h)
	return &orstedrpc.ResultMessage{Result: 0}, err
}

func (s* Server) ViewHostFile(context.Context, *orstedrpc.EmptyMessage) (*orstedrpc.HostList, error) {
	res, err := orsteddb.ViewHostedFileDb()
	for _, h := range res.Hostlist {
		h.Filename = profiles.Config.Endpoints["hostendpoint"] + h.Filename
	}
	return res, err
}

func (s* Server) AddRoute(c context.Context, r *orstedrpc.RouteReq) (*orstedrpc.Route, error) {


	err := autoroute.AddRoute(r.BeaconId, r.Subnet)
	if err != nil {
		fmt.Println("Error Adding Route ", err)
		return nil, err
	}

	var t orstedrpc.TaskReq
	t.BeacondId = r.BeaconId
	t.Command = "autoroute"
	t.PrettyCommand = "Automatic Command: Autoroute"

	s.AddTask(c, &t)
	
	return nil, nil
}

func (s* Server) DeleteRoute(c context.Context, r *orstedrpc.Route) (*orstedrpc.ResultMessage, error) {
	return nil, autoroute.DeleteSubnetRoute(r.BeaconId, r.Subnet)
}

func (s* Server) ListRoute(c context.Context, e *orstedrpc.EmptyMessage) (*orstedrpc.RouteList, error) {
	return autoroute.ListRoute(), nil
}

func (s* Server) AddRevPortFwd(c context.Context, r *orstedrpc.RevPortFwdReq) (*orstedrpc.Route, error) {


	err := autoroute.AddReversePortForwardInRoute(r.BeaconId, r.RemoteSrc, r.LocalDst)
	if err != nil {
		fmt.Println("Error Adding Route ", err)
		return nil, err
	}

	var t orstedrpc.TaskReq
	t.BeacondId = r.BeaconId
	t.Command = "autoroute"
	t.PrettyCommand = "Automatic Command: Autoroute"

	s.AddTask(c, &t)
	
	return nil, nil
}

func (s* Server) DeleteRevPortFwd(c context.Context, r *orstedrpc.Route) (*orstedrpc.ResultMessage, error) {

	return nil, autoroute.DeletePortFwdFromRoute(r.BeaconId, r.Rportfwd)
}
