# sesh

This is a session manager for Go.


There are four main functions:

	sm := sesh.NewSM()//creates session object
	sm.StartSession(w, uniqueString)//w is http.Response Writer, unique string is username or something unique (like whatever you're using to identify/store users in a database)
	sm.VerifySession(r)//r is *http.Request
	sm.EndSession(r)//r is *http.Request


Under the hood:

The sessions are stored as a 2 dimensional slice. Its more specifically a slice
of chains, and each chain holds (by default) 1000 session structs. The position
of the session and the identifier(passed in to StartSession) are stored in the 
returned cookie


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



