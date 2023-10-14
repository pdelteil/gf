# gf

A wrapper around grep to avoid typing common patterns.

## What the fork? 

This repo was abandoned by the great [tomnomnom](https://github.com/tomnomnom). I will try to add some rules and keep it updated (at least the installation instructions) 


## Install

If you've got Go installed and configured you can install `gf` with:

```
go install github.com/pdelteil/gf@latest
```
### Config

Tested with go version `go1.21.3`. If you have an older version try updating using [this](https://github.com/udhos/update-golang).
```
# installation path of module 
export GO111MODULE=on
module_path=$(go list -m -f '{{.Dir}}' github.com/pdelteil/gf@latest)

echo "source $module_path/gf-completion.bash" >> ~/.bashrc

source ~/.bashrc

#copy example pattern files to `~/.gf`

cp -r $module_path/examples/* ~/.gf

```

### First run

```
gf php-errors file

```
### Flags

```
gf -list
gf -dump pattern
gf -save pattern-name flags 'search-pattern'

```

With the flag **dump** you can check the resulting command created by a pattern:

```
gf -dump xss
grep -iE "(q=|s=|search=|lang=|keyword=|query=|page=|keywords=|year=|view=|email=|type=|name=|p=|callback=|jsonp=|api_key=|api=|password=|email=|emailto=|token=|username=|csrf_token=|unsubscribe_token=|id=|item=|page_id=|month=|immagine=|list_type=|url=|terms=|categoryid=|key=|l=|begindate=|enddate=)" .
```

## Why?

When auditing code bases, looking at the output of [meg](https://github.com/tomnomnom/meg), or just generally dealing with large amounts of data. I often end up using fairly complex patterns like this one:

```
▶ grep -HnrE '(\$_(POST|GET|COOKIE|REQUEST|SERVER|FILES)|php://(input|stdin))' *
```

So the above command becomes simply:

```
▶ gf php-sources
```

### Pattern Files

The pattern definitions are stored in `~/.gf` as little JSON files:

```
▶ cat ~/.gf/php-sources.json
{
    "flags": "-HnrE",
    "pattern": "(\\$_(POST|GET|COOKIE|REQUEST|SERVER|FILES)|php://(input|stdin))"
}
```

To help reduce pattern length and complexity, you can specify a list of multiple patterns:

```
▶ cat ~/.gf/php-sources-multiple.json
{
    "flags": "-HnrE",
    "patterns": [
        "\\$_(POST|GET|COOKIE|REQUEST|SERVER|FILES)",
        "php://(input|stdin)"
    ]
}
```

There are some more example pattern files in the `examples` directory.

You can use the `-save` flag to create pattern files from the command line:

```
▶ gf -save php-serialized -HnrE '(a:[0-9]+:{|O:[0-9]+:"|s:[0-9]+:")'
```

### Auto Complete

There's an auto-complete script included, so you can hit 'tab' to show you what your options are:

```
▶ gf <tab>
base64       debug-pages  fw           php-curl     php-errors   php-sinks    php-sources  sec          takeovers    urls
```


### Using custom engines

There are some amazing code searching engines out there that can be a better replacement for grep.
A good example is [the silver searcher](https://github.com/ggreer/the_silver_searcher).
It's faster and presents the results in a more visually digestible manner.
In order to utilize a different engine, add `engine: <other tool>` to the relevant pattern file:
```bash
# Using the silver searcher instead of grep for the aws-keys pattern:
# 1. Adding "ag" engine
# 2. Removing the E flag which is irrelevant for ag
{
  "engine": "ag",
  "flags": "-Hanr",
  "pattern": "([^A-Z0-9]|^)(AKIA|A3T|AGPA|AIDA|AROA|AIPA|ANPA|ANVA|ASIA)[A-Z0-9]{12,}"
}
```
* Note: Different engines use different flags, so in the example above, the flag `E` has to be removed from the `aws-keys.json` file in order for ag to run.

## Contributing
Interested in new pattern files! 
Bug fixes are also welcome as always :)
