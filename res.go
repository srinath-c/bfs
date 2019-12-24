package main

const html = `<html>
<head>
	<title>{{.Title}}</title>
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<style>
	.dirs {
		width : 100%;
		/* background-color: #fbd03e; */
	}
	
	.folder {
		padding : 25px !important;
	}
	
	
	.entry:active {
		box-shadow : #555555 2px 2px 5px;
	}
	
	a, a:active, a:focus {
		outline: none;
	}
	
	.entry {
		outline: none;
		box-shadow : #555555 5px 5px 5px;
		display: inline-block;  
		max-width : 200px;
		padding : 10px;
		color : #999999;
		background-color : #444444;
		margin : 10px;
	}
	
	.filename {
	  text-align: center;
	}
	
	.ext {
		display: none;
	  text-align: center;
	}
	
	a {
	  color : #DDDDDD;
	  text-decoration: none;
	}
	
	h1,h2 {
		padding : 10px;
		font-weight : bold;
	}
	
	body {
		outline: none;
		margin : 0px;
	  font-family: 'Cantarell','URW Gothic', 'Sans Serif', 'Monospace';
	  color : #999999;
	  background-color: #333333;
	  width : 100%;
	  height : 100%;
	  overflow-x: hidden;
	}
	</style>
</head>
<body><center>
<h1>{{.Title}}</h1>
<div class="dirs">
<h2>Folders</h2>
{{range $i, $a := .FileInfos}}
	{{if .IsDir}}
	<a href="{{.Href}}">
		<div class="entry folder">
			<div class="filename">
				{{.Filename}}
			</div>
			<div class="ext">{{.Extension}}</div>
		</div>
	</a>
	{{end}}
{{end}}
</div>
<br>
<div class="files">
<h2>Files</h2>
{{range $i, $a := .FileInfos}}
	{{if .IsDir}}
	{{else}}
	<a href="{{.Href}}">
		<div class="entry">
			<div class="filename">
				{{.Filename}}
			</div>
			<div class="ext">{{.Extension}}</div>
		</div>
	</a>
	{{end}}
{{end}}
</div>
</center>
</body>
</html>`
