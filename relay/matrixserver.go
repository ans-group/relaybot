package relay

import (
	"context"

	log "github.com/sirupsen/logrus"
	"maunium.net/go/mautrix"
)

func init() {
	RegisterServer(func() []Server {
		return loadMatrixServers()
	})
}

type MatrixServerLogger struct{}

func (l *MatrixServerLogger) Debugfln(message string, args ...interface{}) {
	log.Debugf(message, args...)
}

// MatrixServerConfig represents the configuration of a Matrix server
type MatrixServerConfig struct {
	Homeserver  string `mapstructure:"homeserver"`
	Username    string `mapstructure:"username"`
	Password    string `mapstructure:"password"`
	DisplayName string `mapstructure:"display_name"`
}

type MatrixServer struct {
	*ServerBase

	config MatrixServerConfig
	client *mautrix.Client
}

func loadMatrixServers() []Server {
	var serverConfigs map[string]MatrixServerConfig
	getServerConfigs("matrix", &serverConfigs)

	var servers []Server
	for server, config := range serverConfigs {
		servers = append(servers, NewMatrixServer(server, config))
	}

	return servers
}

// NewMatrixServer returns a pointer to an initialised MatrixServer struct
func NewMatrixServer(name string, config MatrixServerConfig) *MatrixServer {
	return &MatrixServer{
		ServerBase: NewServerBase(name),
		config:     config,
	}
}

// Connect connects to the configured Matrix homeserver
func (s *MatrixServer) Connect() error {
	log.Infof("Connecting to %s as %s", s.config.Homeserver, s.config.Username)
	client, err := mautrix.NewClient(s.config.Homeserver, "", "")
	if err != nil {
		return err
	}
	s.client = client

	if Debug {
		s.client.Logger = &MatrixServerLogger{}
	}

	resp, err := s.client.Login(&mautrix.ReqLogin{Type: "m.login.password", User: s.config.Username, Password: s.config.Password})
	if err != nil {
		return err
	}
	s.client.SetCredentials(resp.UserID, resp.AccessToken)

	if len(s.config.DisplayName) > 0 {
		s.client.SetDisplayName(s.config.DisplayName)
	}

	log.Infof("Connected successfully")
	return nil
}

// SetTargets sets Matrix server targets and joins corresponding IRC rooms
func (s *MatrixServer) SetTargets(targets []Target) error {
	s.Targets = targets
	return s.joinTargetRooms()
}

func (s *MatrixServer) joinTargetRooms() error {
	for _, target := range s.Targets {
		err := s.joinTargetRoom(target)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *MatrixServer) joinTargetRoom(target Target) error {
	joinedRoomsResp, err := s.client.JoinedRooms()
	if err != nil {
		return err
	}

	for _, joinedRoom := range joinedRoomsResp.JoinedRooms {
		if joinedRoom == target.Name {
			log.Debugf("Skipping joining room [%s] as already joined", target.Name)
			return nil
		}
	}

	log.Infof("Joining room [%s]", target.Name)
	_, err = s.client.JoinRoom(target.Name, "", nil)
	if err != nil {
		return err
	}

	return nil
}

// Read starts reading from the Matrix server rooms
func (s *MatrixServer) Read(ctx context.Context, readChan chan *TargetMessage) error {
	syncer := s.client.Syncer.(*mautrix.DefaultSyncer)
	syncer.OnEventType(mautrix.EventMessage, func(evt *mautrix.Event) {
		log.Tracef("Got event %+v", evt)

		if evt.Sender == s.client.UserID {
			log.Debugf("Ignoring message as bot user")
			return
		}

		target, err := s.GetTarget(evt.RoomID)
		if err != nil {
			log.Errorf("Failed to retrieve target for event: %s", err.Error())
			return
		}

		displayNameResp, err := s.client.GetDisplayName(evt.Sender)
		if err != nil {
			log.Errorf("Failed to retrieve display name for user [%s]: %s", evt.Sender, err.Error())
			return
		}

		s.client.MarkRead(evt.RoomID, evt.ID)

		log.Tracef("Creating new message for sender [%s] with display name [%s] with content [%s] for event [%s]", evt.Sender, displayNameResp.DisplayName, evt.Content.Body, evt.ID)
		msg := NewTargetMessage(target, Target{}, TargetMessagePayload{User: displayNameResp.DisplayName, Msg: evt.Content.Body})

		readChan <- msg
	})

	go s.client.Sync()

	select {
	case <-ctx.Done():
		s.client.StopSync()
		return ctx.Err()
	}
}

// Write writes TargetMessage msg to target destination
func (s *MatrixServer) Write(msg *TargetMessage) error {
	log.Debugf("Sending message [%s] to room [%s]", msg.GetMessage(), msg.Destination.Name)
	_, err := s.client.SendText(msg.Destination.Name, msg.GetMessage())
	return err
}
