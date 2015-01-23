The code behind [@optimistic_dev](https://twitter.com/optimistic_dev).

Everything is awesome, let's tell people how awesome their github projects are!

Introduction
============

This twitter bot keeps track of links to github projects, and when a given project has been tweeted
five times, tells everyone how awesome the project is.

Contributing
============

I'd love improvements to the code, here are the steps you need to go through to make it work.

1. Set up a [Go development envrionment](https://golang.org/doc/install).
1. Fork this repo on github.
1. Download your copy of the code.

    go get github.com/YOUR_NAME/optimistic_dev


1. Create your own app on [apps.twitter.com](https://apps.twitter.com/), make sure it has read-write permissions.
1. Create a file called .env in `$GOPATH/src/github.com/YOUR_NAME/optimistic_dev/.env` that contains the following content based on your app settings:

    TWITTER_CONSUMER_KEY=<...>
    TWITTER_CONSUMER_SECRET=<...>
    TWITTER_ACCESS_KEY=<...>
    TWITTER_ACCESS_SECRET=<...>

1. Build the app

    go build

1. Run it

    ./optimistic_dev

1. Make any code changes you'd like, and test that they work.
1. Commit your changes
1. Push them to your fork
1. Send me a pull request

License
=======

optimistic_dev is licensed under the MIT license, see LICENSE.MIT for details.
