package handlers

import (
	"fmt"
	codecservices "github.com/stretchrcom/codecs/services"
	"github.com/stretchrcom/goweb/webcontext"
	"net/http"
	"strings"
)

type HttpHandler struct {

	// codecServices is the codec service object to use to go from bytes to objects
	// and vice versa.
	codecService codecservices.CodecService

	// Handlers represent a pipe of handlers that will be used
	// to handle requests.
	Handlers Pipe

	// errorHandler represents the Handler that will be used to handle errors.
	errorHandler Handler
}

func NewHttpHandler(codecService codecservices.CodecService) *HttpHandler {
	h := new(HttpHandler)

	// make pre, process and post pipes
	h.Handlers = make(Pipe, 3)
	h.Handlers[0] = make(Pipe, 0) // pre
	h.Handlers[1] = make(Pipe, 0) // process
	h.Handlers[2] = make(Pipe, 0) // post

	h.codecService = codecService

	return h
}

func (handler *HttpHandler) CodecService() codecservices.CodecService {
	return handler.codecService
}

// ServeHTTP servers the 
func (handler *HttpHandler) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {

	// make the context
	ctx := webcontext.NewWebContext(responseWriter, request, handler.codecService)

	// run it through the handlers
	_, err := handler.Handlers.Handle(ctx)

	// do we need to handle an error?
	if err != nil {

		// set the error
		ctx.Data().Set("error", err)

		// tell the handler to handle it
		handler.ErrorHandler().Handle(ctx)

	}

}

// ErrorHandler gets the Handler that will be used to handle errors.
//
// If no error handler has been set, a default error handler will be returned
// which will just write the error out in plain text.  If you are building an API,
// it is recommended that you roll your own ErrorHandler.
//
// For more information on rolling your own ErrorHandler, see the SetErrorHandler
// method.
func (h *HttpHandler) ErrorHandler() Handler {

	if h.errorHandler == nil {

		h.errorHandler = &DefaultErrorHandler{}

	}

	return h.errorHandler
}

// SetErrorHandler sets the Handler that will be used to handle errors.
//
// The error handler is like a normal Handler, except with a few oddities. 
// The WillHandle method never gets called on the ErrorHandler, and any errors
// returned from the Handle method are ignored (as is the stop argument).
// If you want to log errors, you should do so from within the ErrorHandler.
//
// Goweb will place the error object into the context.Data().Get("error") map.
func (h *HttpHandler) SetErrorHandler(errorHandler Handler) {
	h.errorHandler = errorHandler
}

// HandlersPipe gets the pipe for handlers.
func (h *HttpHandler) HandlersPipe() Pipe {
	return h.Handlers[1].(Pipe)
}

// PreHandlersPipe gets the handlers that are executed before processing begins.
func (h *HttpHandler) PreHandlersPipe() Pipe {
	return h.Handlers[0].(Pipe)
}

// PostHandlersPipe gets the handlers that are executed after processing completes.
func (h *HttpHandler) PostHandlersPipe() Pipe {
	return h.Handlers[2].(Pipe)
}

// AppendHandler appends a handler to the processing pipe.
func (h *HttpHandler) AppendHandler(handler Handler) {
	h.Handlers[1] = h.HandlersPipe().AppendHandler(handler)
}

// AppendPreHandler appends a handler to be executed before processing begins.
func (h *HttpHandler) AppendPreHandler(handler Handler) {
	h.Handlers[0] = h.PreHandlersPipe().AppendHandler(handler)
}

// PrepentPreHandler prepends a handler to be executed before processing begins.
func (h *HttpHandler) PrependPreHandler(handler Handler) {
	h.Handlers[0] = h.PreHandlersPipe().PrependHandler(handler)
}

// AppendPostHandler appends a handler to be executed after processing completes.
func (h *HttpHandler) AppendPostHandler(handler Handler) {
	h.Handlers[2] = h.PostHandlersPipe().AppendHandler(handler)
}

// PrependPostHandler prepends a handler to be executed after processing completes.
func (h *HttpHandler) PrependPostHandler(handler Handler) {
	h.Handlers[2] = h.PostHandlersPipe().PrependHandler(handler)
}

/*
	Debug and information
*/

// String generates a list of the handlers registered inside this HttpHandler.
func (h *HttpHandler) String() string {
	return stringForHandlers(h.Handlers, 0)
}

// stringForHandlers generates the string for the handlers array indented to the
// appropriate level.
func stringForHandlers(handlers []Handler, level int) string {

	lines := []string{}
	var levelStr string = strings.Repeat("  ", level)

	for handlerIndex, handler := range handlers {
		if pipe, ok := handler.(Pipe); ok {
			lines = append(lines, fmt.Sprintf("%sPipe %d:", levelStr, handlerIndex))
			lines = append(lines, stringForHandlers(pipe, level+1))
		} else {
			lines = append(lines, fmt.Sprintf("%s%s", levelStr, handler))
		}
	}

	return strings.Join(lines, "\n")

}
