# Smart Golang Error Handling

This blog post is based on a great [GopherCon 2016 talk](https://www.youtube.com/watch?v=lsBF58Q-DnY) by Dave Cheney.

TODO:
- introduction
- returning errors vs passing errors with extra data
- pkg/errors to the rescue
    - wrapping errors,
    - causing errors,
    - stack traces
- example line-by-line (bare, concat, wrap)
    - string message,
    - error type,
    - original error,
    - original type,
    - stack trace
