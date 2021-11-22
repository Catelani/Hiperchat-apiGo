package main

import (
	"github.com/Mateus-pilo/go-whats-opt/ctl"
	"github.com/Mateus-pilo/go-whats-opt/hlp/auth"
	"github.com/Mateus-pilo/go-whats-opt/hlp/router"
)

// Initialize Function in Main Route
func init() {
	// Set Endpoint for Root Functions
	router.Router.Get(router.RouterBasePath, ctl.GetIndex)
	router.Router.Get(router.RouterBasePath+"/health", ctl.GetHealth)

	// Set Endpoint for Authorization Functions
	router.Router.With(auth.Basic).Get(router.RouterBasePath+"/auth", ctl.GetAuth) 

	// Set Endpoint for WhatsApp Functions
	router.Router.With(auth.JWT).Post(router.RouterBasePath+"/login", ctl.WhatsAppLogin)
	router.Router.With(auth.JWT).Post(router.RouterBasePath+"/send/text", ctl.WhatsAppSendText)
	router.Router.With(auth.JWT).Post(router.RouterBasePath+"/send/document", ctl.WhatsAppSendDocument)
	router.Router.With(auth.JWT).Post(router.RouterBasePath+"/send/image", ctl.WhatsAppSendImage)
	router.Router.With(auth.JWT).Post(router.RouterBasePath+"/send/audio", ctl.WhatsAppSendAudio)
	router.Router.With(auth.JWT).Post(router.RouterBasePath+"/logout", ctl.WhatsAppLogout)
}

