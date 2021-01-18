package logger

import (
	"sync"
)

type Log struct {
	AddMessage func(Message)
	Warnings   func() []Message
	Errors     func() []Message
	Done       func() []Message
}

type MessageKind uint8

const (
	Error MessageKind = iota
	Warning
)

type Message struct {
	Kind  MessageKind
	Data  MessageData
}

type MessageData struct {
	Text string
}

func (kind MessageKind) String() string {
	switch kind {
	case Error:
		return "error"
	case Warning:
		return "warning"
	default:
		panic("Unknown message kind")
	}
}

func New() Log {
	var msgs []Message
	var mutex sync.Mutex

	return Log{
		AddMessage: func(msg Message) {
			mutex.Lock()
			defer mutex.Unlock()
			msgs = append(msgs, msg)
		},
		Warnings: func() []Message {
			mutex.Lock()
			defer mutex.Unlock()
			warnings := []Message{}
			for _, msg := range msgs {
				if msg.Kind == Warning {
					warnings = append(warnings, msg)
				}
			}
			return warnings
		},
		Errors: func() []Message {
			mutex.Lock()
			defer mutex.Unlock()
			errors := []Message{}
			for _, msg := range msgs {
				if msg.Kind == Warning {
					errors = append(errors, msg)
				}
			}
			return errors
		},
		Done: func() []Message {
			mutex.Lock()
			defer mutex.Unlock()
			return msgs
		},
	}
}

func (log Log) AddError(text string) {
	log.AddMessage(Message{
		Kind: Error,
		Data: MessageData{Text: text},
	})
}

func (log Log) AddWarning(text string) {
	log.AddMessage(Message{
		Kind: Warning,
		Data: MessageData{Text: text},
	})
}
