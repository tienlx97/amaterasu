package handling

import (
	v2messaging "amaterasu/cqrs/infrastructure/v2.messaging"
	"context"
)

/*
In the V2 release, the system used the Windows Azure Service Bus to
deliver all commands to their recipients. This meant that the system
delivered the commands asynchronously. In the V3 release, the MVC
controllers now send their commands synchronously and in-process
in order to improve the response times in the UI by bypassing the
command bus and delivering commands directly to their handlers. In
addition, in the ConferenceProcessor worker role, commands sent to
Order aggregates are sent synchronously in-process using the same
mechanism.
*/
type SyncCommandBus struct {
	bus        v2messaging.CommandBus
	dispatcher *CommandDispatcher
}

func New_SyncCommandBus(bus v2messaging.CommandBus) {
	syncCb := &SyncCommandBus{}
	syncCb.bus = bus
	syncCb.dispatcher = New_CommandDispatcher()
}

func (scb SyncCommandBus) Register(commandHandler v2messaging.ICommandHandler) error {
	return scb.dispatcher.Register(commandHandler)
}

func (scb SyncCommandBus) Send(ctx context.Context, command v2messaging.Envelope) error {
	if !scb.doSend(command) {
		// TODO trace log
		// Trace.TraceInformation("Command with id {0} was not handled locally. Sending it through the bus.", command.Body.Id);
		return scb.bus.Send(ctx, command)
	}
	return nil
}

func (scb SyncCommandBus) Sends(ctx context.Context, commands ...v2messaging.Envelope) error {
	// pending command
	for len(commands) > 0 {
		if scb.doSend(commands[0]) {
			commands = commands[1:]
		} else {
			break
		}
	}

	// Commands were not handled locally. Sending it and all remaining commands through the bus
	if len(commands) > 0 {
		// TODO trace log
		scb.bus.Sends(ctx, commands...)
	}

	return nil
}

func (scb SyncCommandBus) doSend(command v2messaging.Envelope) bool {
	// TODO trace log
	return scb.dispatcher.ProcessMessage(command) == nil
}