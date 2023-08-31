## smstool

A simple command line tool using which you can skip phone number based SMS
verification by using a temporary phone number that acts like a proxy.

**Latest update:** The tool no longer uses upmasked.com, as the service went
down. We are using another provider which provides more phone numbers across
more countries. Make sure you pull the main branch before compiling.

this is a fork/quick cleanup of: [github.com/Narasimha1997/fake-sms](https://github.com/Narasimha1997/fake-sms)

### Features:

- CLI, which is easier to use. Provides a local file based DB to save and
- manage a list of fake phone numbers to help you remember and reuse.

### To build:

The build process is simple, it is just like building any other Go module.
Follow the steps below:

```sh
go install github.com/moistari/smstool@latest
```

#### Steps to use:

1. Register a number in local DB: You can register a number by selecting one of
   the available numbers as shown below.

![register-number](./gifs/add.gif)

2. Get the messages from any registered number: You can select a number which
   was saved in step-1 and view its messages as a list. The tool will also save
   the dump as json in the format `${PWD}/selected-phone-number.json`. As shown
   below:

![get-messages](./gifs/messages.gif)

3. Optionally, you can choose to delete the rembered numbers or list them.

#### Acknowledgements

The similar tool is also available in pure shell script. [Check this
out.](https://github.com/sdushantha/tmpsms)

#### Contributing

The tool is very simple and I don't think there is any major feature missing.
But I would welcome any kind of suggestion, enhancements or a bug-fix from the
community. Please open an issue to discuss or directly make a PR!!
