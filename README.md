# wtfamisending 

#####http://wtfamisending.com/  
See your ****ing request, HTTP headers and all that buzz.  
Maybe this can help you debug stuff, or this may just help you see what you are really sending to all those pesky websites.

#### Config (filename: "config"):
```
; Database 
[DB]
UserName = SomeUserName
Password = Somepassword
Name = SomeDBName

; Server 
[Server]
Port = :5000
```

####How it works:
```
req wtfamisending.com 
	=> redirect to: wtfamisending.com/:id?html=true
		and body of :id

req wtfamisending.com/:id?html=true 
	=> show html pretty page

req wtfamisending.com/:id 
	=> show raw request
```