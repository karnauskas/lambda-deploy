# lambda-deploy

## A tool to push lambda functions to a specific environment

## The cli tool assumes a directory structure:

    - dir
    
    -- deploy.json

    -- src/
    
    --- someFile.py
    
    --- config.json


## Usage
    lambda-deploy -env=dev -directory=/Users/alex/Code/Python/my-lambda
    
Assuming my-lambda contains a deploy.json file as well as a src folder, everything should work as intended.