package relay

import (
	"context"
	"crypto/tls"

	log "github.com/sirupsen/logrus"
	irc "github.com/thoj/go-ircevent"
)

const (
	EventWelcome = "001" // RPL_WELCOME
	EventPrivMsg = "PRIVMSG"
)

func init() {
	RegisterServer(func() []Server {
		return loadIRCServers()
	})
}

// IRCServerConfig represents the configuration of an IRC server
type IRCServerConfig struct {
	Host          string `mapstructure:"host"`
	UseTLS        bool   `mapstructure:"use_tls"`
	SkipTLSVerify bool   `mapstructure:"skip_tls_verify"`
	Username      string `mapstructure:"username"`
	Password      string `mapstructure:"password"`
	Nick          string `mapstructure:"nick"`
}

type IRCServer struct {
	*ServerBase

	config IRCServerConfig
	conn   *irc.Connection
}

func loadIRCServers() []Server {
	var serverConfigs map[string]IRCServerConfig
	getServerConfigs("irc", &serverConfigs)

	var servers []Server
	for ircServer, config := range serverConfigs {
		servers = append(servers, NewIRCServer(ircServer, config))
	}

	return servers
}

// NewIRCServer returns a pointer to an initialised IRCServer struct
func NewIRCServer(name string, config IRCServerConfig) *IRCServer {
	return &IRCServer{
		ServerBase: NewServerBase(name),
		config:     config,
	}
}

// Connect connects to the configured IRC homeserver
func (s *IRCServer) Connect() error {
	s.conn = irc.IRC(s.config.Nick, s.config.Username)
	s.conn.Password = s.config.Password
	s.conn.UseTLS = s.config.UseTLS

	if s.config.SkipTLSVerify {
		s.conn.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}

	if Debug {
		s.conn.VerboseCallbackHandler = true
		s.conn.Debug = true
	}

	err := s.conn.Connect(s.config.Host)
	if err != nil {
		return err
	}

	welcomeDone := false
	s.conn.AddCallback(EventWelcome, func(*irc.Event) {
		welcomeDone = true
	})

	for {
		if welcomeDone {
			return nil
		}
	}
}

// SetTargets sets IRC server targets and joins corresponding IRC rooms
func (s *IRCServer) SetTargets(targets []Target) error {
	s.Targets = targets
	for _, target := range targets {
		log.Infof("Joining room [%s]", target.Name)
		s.conn.Join(target.Name)
	}
	return nil
}

// Read starts reading from the IRC server rooms
func (s *IRCServer) Read(ctx context.Context, readChan chan *TargetMessage) error {
	s.conn.AddCallback(EventPrivMsg, func(evt *irc.Event) {
		log.Tracef("Got event %+v", evt)

		if evt.Nick == s.config.Nick {
			log.Debugf("Ignoring message as bot user")
			return
		}

		target, err := s.GetTarget(evt.Arguments[0])
		if err != nil {
			log.Errorf("Failed to retrieve target for event: %s", err.Error())
			return
		}

		log.Tracef("Creating new message for sender [%s] with content [%s]", evt.Nick, evt.Message())
		readChan <- NewTargetMessage(target, Target{}, TargetMessagePayload{User: evt.Nick, Msg: evt.Message()})
	})

	go s.conn.Loop()

	select {
	case <-ctx.Done():
		s.conn.Quit()
		return ctx.Err()
	}
}

// Write writes TargetMessage msg to target destination
func (s *IRCServer) Write(msg *TargetMessage) error {
	log.Debugf("Sending message [%s] to room [%s]", msg.GetMessage(), msg.Destination.Name)
	s.conn.Privmsgf(msg.Destination.Name, msg.GetMessage())
	return nil
}
