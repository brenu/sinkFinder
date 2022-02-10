# :mag_right: SinkFinder

Tool that receives URLs of JavaScript files and analyses them in order to spot possible sinks that may lead to DOM-Based XSS cases.

## :gear: Installation

If you've got Go installed and configured, you can install SinkFinder using the command below:

```console
foo@bar:~$ go install github.com/brenu/sinkFinder@latest
```

## :arrow_forward: Usage

The usage of SinkFinder consists of passing a list of domains by using the pipe, just like the examples below. Let's say you have a list of .js links and want to analyse them. In order to do so, you just have to call sinkFinder this way:

```console
foo@bar:~$ cat aliveJsFiles.txt | sinkFinder
```

By default, SinkFinder just executes one HTTP request per second. In case you would like to change this rate limiting, you may use the `-r` flag for that, just like this:

```console
// Setting the maximum number of HTTP requests per second to 150
foo@bar:~$ cat aliveJsFiles.txt | sinkFinder -r 150
```

Also, by default, SinkFinder doesn't work concurrently. In case you would like to make it work in multiple "threads", you may use the `-t` flag for that, just like this:

```console
// Setting the number of concurrent threads to 150
foo@bar:~$ cat aliveJsFiles.txt | sinkFinder -t 100
```
It's also possible to write any results into a file by using the `-o` flag. It's important to mention that it uses an append approach, so that whenever you choose a file with content inside, it will not erase it. Here's an example of using this flag:
```console
// Setting 'results' as the output file
foo@bar:~$ cat aliveJsFiles.txt | sinkFinder -o results
```

## :balance_scale: Disclaim :mag_right:

Use it with caution. You are responsible for your actions. I assume no liability and I'm not responsible for any misuse or damage.