package dbsync

// client:
// recieve db if first time
// copy true db
// change the copy and log changes (into ram or logfile on exit)
// replay server patches on the original, and then ours without logging
//
// server:
// send initial db if first time
// if have patches since last time, send them
// and log last patch sent
// else play their patches and log them
// and last patch sent to them would be theirs lol
//
// 1. how to send and recieve sqlite files
// 2. how to copy the file
// 3. how to log patches in a replayable way
// 4. how to replay patches
// 5. how to send and recieve patches
// 6. write server and client logic
