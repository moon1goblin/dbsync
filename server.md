# how to server

### can we just send the sqlite files over https lol?
our db woulndt be very big, like a few mb max and thats with a looot of events
so apparently yes we can and we should try and benchmark

actually were gonna need to send the file anyway, at least initally
but not complicating logging changes and just sending sqlite files over is an explorable idea

### or, send the changes
on the client: after getting the inital file
, we store the changes (patches) as a log and send just those

- how do we know what the change was?
store last patch id we sent that client on the server
the server looks at last patch it sent us and just sends all the patches since then

- we apply changes to our local db untill theres internet duh
then its these 3 scenarios:

#### A. client has changes, server is the same

just send the patches to the server, it applies them and were guchi

#### B. client is the same, server has changes

just get the patches from the server and apply them on the client bam all good

#### C. client and server are both different, skull emoji

the servers db is the authority, thats the only way to do it yea

first get the patches from the server
, roll back the db to last true state
, and apply those

then we try to apply ours somehow
- if theres no conflict then were good
- if server added line and we add the same line, then just dont do shit, everyones happy

- if server changed line and we didnt, thats a buisness logic decision
, who wins can be decided on who was the last to change it or something
- if server deleted line and we changed it
, then we are fucked.
can we somehow ressurect it? so like, have a column in db that says deleted but dont delete the line
wait we cant change events in our app anyways, we delete them and create the same event but different
well that solves the ressurection problem

, and then its just scenario A, send our patches to the server and were done

after applying, erase the logs because we dont need those anymore

### how do we send the initial db?
i think we can just send the binary file lol, its looks the same on all platforms
and binary is the most efficient
i dont think we need to compress it, its a really small db, but maybe in the future

### what would a patch even look like?
can we just save all sql.execs we do along with timestamps
wait this is exactly what we need:

https://github.com/simukti/sqldb-logger

and so yea we can just send the json {sql string, timestamp}
is using sql strings like that even safe? dependency injection hello

### server side, what do we do?
execute patches and log them with {id, sql string, timestamp} :)


edit: fuck no it is not that simple

would be nice to somehow have a unified db view
that we record execs from the last true version
and we can immidiatly roll back to that version

but for that well have to rewrite sqlites vfs and who knows what else lmao

so we have the original true file AND the file we make changes to
(i call it true db file because the server actually
knows what it is and knows the difference since then)
how slow is it to copy the true db every time were resync?
AND we also store the exec log that is the difference between these two dbs
(computing vs storing it is another thing but fuck not now)

idk lets just try this shit?
and then see how bad it is

so
1. somehow store execs and reproduce them
2. send sqlite files in binary over http

3. everything else
