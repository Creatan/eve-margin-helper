# Eve Margin Helper 
Station trading profit margin calculator for windows

## Usage

Eve margin helper takes two commandline parameters  
`-broker` for broker's fee  
`-tax` for sales tax  

Example:
```
emh -broker 2.77 -tax 2
```

You can also issue a `clean` command for removing old log files
```
emh clean
```

## macOS and Linux

Install Go and build with 
`go build -o <outputfile>`
