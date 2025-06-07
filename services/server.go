package services

import (
	"net/http"
	"time"

	"github.com/zefrenchwan/scrutateur.git/engines"
	"github.com/zefrenchwan/scrutateur.git/storage"
)

// LOCAL_RESOURCES_PATH is the path of all resources to load
const LOCAL_RESOURCES_PATH = "/app/static"

// Init is the place to add all links endpoint -> handlers
func Init(dao storage.Dao, secret string, tokenDuration time.Duration) engines.ProcessingEngine {
	server := engines.NewProcessingEngine(dao)

	// technical endpoint to prove app is up
	server.AddProcessors("GET", "/status", func(context *engines.HandlerContext) error { context.Build(http.StatusOK, "", nil); return nil })

	// login is the connection handler
	loginHandler := engines.BuildLoginHandler(secret, tokenDuration)
	server.AddProcessors("POST", "/login", loginHandler)

	//////////////////////////////////
	// STATIC UNPROTECTED RESOURCES //
	//////////////////////////////////
	staticHandler := engines.BuildStaticHandler("/app/static/", LOCAL_RESOURCES_PATH)
	paths, err := engines.LoadStaticResources(LOCAL_RESOURCES_PATH)
	if err != nil {
		panic(err)
	}

	for _, path := range paths {
		server.AddProcessors("GET", path, staticHandler)
	}

	/////////////////////
	// PROTECTED PAGES //
	/////////////////////
	connectionMiddleware := engines.AuthenticationMiddleware(secret, tokenDuration)

	// PAGES FOR AT LEAST A ROLE
	roleValidationMiddleware := engines.RolesBasedMiddleware()

	//////////////////////////////////////////////////////////////////////////////
	// GROUP SELF: USERS GET THEIR OWN INFORMATION OR CHANGE THEIR OWN PASSWORD //
	//////////////////////////////////////////////////////////////////////////////
	server.AddProcessors("GET", "/self/user/whoami", connectionMiddleware, roleValidationMiddleware, endpointUserInformation)
	server.AddProcessors("POST", "/self/user/password", connectionMiddleware, roleValidationMiddleware, engines.EndpointChangePassword)

	/////////////////////////////////////////////
	// GROUP MANAGEMENT: DEAL WITH USER ACCESS //
	/////////////////////////////////////////////
	server.AddProcessors("POST", "/manage/user/create", connectionMiddleware, roleValidationMiddleware, engines.EndpointAdminCreateUser)
	server.AddProcessors("DELETE", "/manage/user/{username}/delete", connectionMiddleware, roleValidationMiddleware, engines.EndpointRootDeleteUser)
	server.AddProcessors("GET", "/manage/user/{username}/access/list", connectionMiddleware, roleValidationMiddleware, engines.EndpointAdminListUserRoles)
	server.AddProcessors("PUT", "/manage/user/{username}/access/edit", connectionMiddleware, roleValidationMiddleware, engines.EndpointAdminEditUserRoles)

	////////////////////////////////
	// END OF HANDLER DEFINITIONS //
	////////////////////////////////
	return server
}
