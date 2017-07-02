This is abstract session manager for Go.


There are four main functions:

------------------------------------------------------------
sm := sesh.NewSessionManager() //creates new session object

c := sm.StartSessionCookie(identifier) // starts the session(logs them in), returns a cookie

err := sm.VerifySessionCookie(c) //takes the cookie, verifies the session(checks theyre 
								 //logged in). Returns nil if Verified
err := sm.EndSessionCookie(c) //takes the cookie, ends the session (logs them out). returns
							  //an error if the cookie was invalid (meaning theyre not logged
							  //in anyways)

------------------------------------------------------------




Some general, in-practice useage:

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



Under the hood:

The sessions are stored as a 2 dimensional slice. Its more specifically a slice
of chains, and each chain holds (by default) 1000 session structs. 


Resizing is done when 3/4 of the last chain is full, and adds one more chain to the 
chains slice, then copies over the contents



How ended sessions arent wasted: They are strung together, like a linked list. 
The position of the first ended and last ended sessions are stored in the SessionManager 
object. The session struct also has the position of another session in it (for use in 
stringing together ended-sessions). so, as sessions are ended, the spot of the newly-ended 
will get stored in it the SessionsManagers last ended, then the SessionManager's lastEnded
vairable will be set to that newly-ended one. 

Now, when StartSession requests the next spot from sm.nextSpot(), next spot does some
analyzing on if that string is up to date or not, and if there is ended spots, it will
send from the bottom and update it sm's openSpot. Otherwise, it sends from the top of 
the last chain.



