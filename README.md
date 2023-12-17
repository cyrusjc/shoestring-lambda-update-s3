# shoestring-go-lambd

This is the repository for the the go lambda function that updates the menu for shoestring.cafe found: https://github.com/cyrusjc/shoestringv2

This piece performs a GET to google sheets API, and parses the data into a .json and uploads the .json files into the S3 bucket of shoestring-cafe-react which allows changing of data on shoestring.cafe/Menu.

Github actions is used for CI/CD and automatically updates the lambda function responsible for the GET request.

#TODO

Github actions to install GO and build on the container vs pushing a .zip onto the repository and uploading to S3 via AWS CLI
