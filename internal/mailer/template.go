//nolint:gosec
package mailer

type TemplateKey string

const (
	TemporaryPasswordKey TemplateKey = "temporaryPassword"
	PasswordRecoveryKey  TemplateKey = "passwordRecovery"
	TitleTextKey         TemplateKey = "titleText"
)

func (t TemplateKey) String() string {
	return string(t)
}

var mailTemplates = map[TemplateKey]string{
	TitleTextKey:         titleTextTmpl,
	PasswordRecoveryKey:  passwordRecoveryTmpl,
	TemporaryPasswordKey: temporaryPasswordTmpl,
}

type TitleText struct {
	Title string
	Text  string
}

const titleTextTmpl = `
<!doctype html>
<html>
<head>
    <meta http-equiv="Content-Type" content="text/html; charset=utf-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <link href="https://fonts.googleapis.com/css?family=Lato:400,300" rel="stylesheet" type="text/css">
    <title>{{.Title}}</title>
    <meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1">
</head>

<body>
<center>
    <div style="padding: 55px 0; width: 800px; text-align: left">
        <h2 style="text-align: left">{{.Title}}</h2>
          
            <p>{{.Text}}</p>
      
    </div>
</center>
</body>
</html>
`

type PasswordRecovery struct {
	Email    string
	Password string
}

const passwordRecoveryTmpl = `
<!doctype html>
<html>
<head>
    <meta http-equiv="Content-Type" content="text/html; charset=utf-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <link href="https://fonts.googleapis.com/css?family=Lato:400,300" rel="stylesheet" type="text/css">
    <title>Password recovery</title>
    <meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1">
</head>

<body>
<center>
    <div style="padding: 55px 0; width: 800px; text-align: left">
        <h2 style="text-align: left">Password recovery. Your new credentials :</h2>
        <p>Email: {{.Email}}</p>
        <p>Password: {{.Password}}</p>
    </div>
</center>
</body> 
</html>
`

type TemporaryPassword struct {
	Email    string
	Password string
}

const temporaryPasswordTmpl = `
<!doctype html>
<html>
<head>
    <meta http-equiv="Content-Type" content="text/html; charset=utf-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <link href="https://fonts.googleapis.com/css?family=Lato:400,300" rel="stylesheet" type="text/css">
    <title>Temporary password</title>
    <meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1">
</head>

<body>
<center>
    <div style="padding: 55px 0; width: 800px; text-align: left">
        <h2 style="text-align: left">Your temporary credentials: </h2>
        <p>Email: {{.Email}}</p>
        <p>Password: {{.Password}}</p>
    </div>
</center>
</body>
</html>
`
