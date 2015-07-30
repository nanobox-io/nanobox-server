/*
Package logtap is a system of log collectors and log drains

Collectors

There is only one collector at the moment but that will be expanded in the
future but the purpose of these collectors is to collect logging data and
Drop that data on a channel the one collector now lists for syslog messages
on a given port.

Drains

Any drain takes the messages collected by a collector and sends them in some
way to there intended functionaity. This can be sending it to stdout, storing them
for later recalling like the historical drain does, or publishng them to another package.


*/
package logtap
