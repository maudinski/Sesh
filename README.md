This is abstract session manager for Go.


There are four main functions:

sm := sesh.NewSessionManager() //creates new session object

c := sm.StartSessionCookie(identifier) // starts the session(logs them in), returns a cookie

err := sm.VerifySessionCookie(c) //takes the cookie, verifies the session(checks theyre 
								 //logged in). Returns nil if Verified

err := sm.EndSessionCookie(c) //takes the cookie, ends the session (logs them out). returns
							  //an error if the cookie was invalid (meaning theyre not logged
							  //in anyways)


Useage:

------------------------------------------------------------
to create a new SessionManager object:
	var sm *sesh.SessionManager // globally
		
	...

	sm = sesh.NewSessionManager() // in main function, before calling http.ListenAndServe()

------------------------------------------------------------
to log someone in (create a session):
	
	identifier := "(username/something unique)"
	c := sm.StartSessionCookie(identifier) // returns a cookie with Name: "session"
	http.SetCookie(w, c) // w is the http.ResponseWriter

------------------------------------------------------------
to check if someone is logged in (verify a session):
	
	c, err := r.Cookie("session") //r is *http.Request
	if err != nil{
		//they dont have that cookie(theyre not logged in/deleted their cookies)	
	}

	err = sm.VerifySessionCookie(c)
	if err != nil{
		//this means they are not logged in/the cookie is invalid	
	}

------------------------------------------------------------
to log someone out (end a session):

	c, err := r.Cookie("session")
	if err != nil {
		//they dont have that cookie
	}

	err = sm.EndSessionCookie(c)
	if err != nil{
		//was an invalid cookie value/ no session anyways	
	}

-----------------------------------------------------------
