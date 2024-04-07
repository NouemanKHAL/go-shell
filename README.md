# Coding Challenge: Build Your Own Shell

This is my solution to the coding challenge [John Crickett's Coding Challenges](https://codingchallenges.fyi/challenges/challenge-shell/).


## Setup

1. Clone the repo
1. Run the tool using of the following approaches:

    ```shell
    # Run the tool automatically using the go command
    $ go run .

    # Build a binary and run it manually
    $ go build -o gosh
    $ ./gosh

    # Install the binary in your environment, and run it:
    $ go install
    $ go-shell
    ```
1. Done!


## Examples

```shell
gosh > $ ls
README.md  go.mod  go.sum  internal  main.go
gosh > $ ls -a
.  ..  .git  README.md	go.mod	go.sum	internal  main.go
gosh > $ pwd
/Users/noueman.khalikine/.noueman/coding-challenges/go-shell
```
