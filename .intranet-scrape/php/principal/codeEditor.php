<?php
/**
 * Created on 21/11/2008 - 15/05/2013
 * 
 * Luis M. Melgratti
 */


$file	= $_GET['file'];
$actualFile = $file;
$subdir = $_GET["dir"];

$path='../';
include ("./autoload.php");
include_once ("../funciones/conexion.php");
include_once ("./sessionCheck.php");


$xmlPath=$datosbase->xmlPath;
$dirXML = '../database/'.$xmlPath.'xml/';

if ($subdir != '')
	$dirXML2 = $dirXML.$subdir.'/';
else
	$dirXML2 = $dirXML;


$error ='';
if ( !is_readable($dirXML2.$file)){
    $error = 'Error de permisos al leer el archivo: '.$dirXML2.$file;
}

if ( !is_file($dirXML2.$file)){
    $error = 'No existe el archivo: '.$dirXML2.$file;
}

if ($error == '')
    $fileContent = file_get_contents($dirXML2.$file);


    ?>
<html>
<head>
<title>EditArea Test</title>

 <link rel="stylesheet" type="text/css" href="../funciones/concat.php?type=css" />
            <?php

            $cssFiles[] = '../css/histrix.css';
            //$cssFiles[] = '../css/user.css.php';

            $cssFiles[] = '../css/smoothness/jquery-ui-1.8.16.custom.css'; //jquery UI
            if ($_SESSION['mobile'] == 1){
                $cssFiles[] = '../css/mobile.css';
             }

            foreach($cssFiles as $n => $cssF) {
                if(is_file($cssF)) {
                    $css_size = filesize($cssF);
                    echo '<link rel="stylesheet" type="text/css" href="'.$cssF.'?'.$css_size.'.css" />';
                    echo "\n";
                }
            }
            ?>


<style type="text/css" media="all">

ul {
    padding-left:0.7em;
    list-style: none;
    border-left:1px solid #ccc;
}
ul li{
padding:3px;
cursor:pointer;
}
.gris {
color:#888;
}

.page {
	/*height:95%;
	width:100%;*/
    bottom:10px;
    top:0px;

}
.msg {
padding:2px;
background-color:red;
color:#FFF;
z-index:99;
margin:0;
}
.CodeMirror-line-numbers {
width: 2.2em;
color: #aaa;
background-color: #eee;
text-align: right;
padding-right: .3em;
font-size: 10pt;
font-family: monospace;
padding-top: .4em;
}

</style>

<script language="javascript" type="text/javascript" src="../lib/codemirror/lib/codemirror.js"></script>
<script language="javascript" type="text/javascript" src="../lib/codemirror/mode/xml/xml.js"></script>
<script language="javascript" type="text/javascript" src="../lib/codemirror/mode/sql/sql.js"></script>
<script language="javascript" type="text/javascript" src="../lib/codemirror/addon/edit/closetag.js"></script>
<script language="javascript" type="text/javascript" src="../javascript/jquery-1.7.2.min.js"></script>
<script language="javascript" type="text/javascript" src="../javascript/jquery-ui-1.8.16.custom.min.js"></script>
 <link rel="stylesheet" type="text/css" href="../lib/codemirror/lib/codemirror.css"/>

</head>
<body>
<?php
/*
<script language="javascript" type="text/javascript" src="../lib/codemirror/js/mirrorframehistrix.js"></script>

*/

if ($_SESSION['EDITOR'] != 'editor'){
    $error = '<div class="error">Must be editor</div>';
    die($error);
}

?>    
<div class="page"  style="width:20%;left:0px;position:absolute; overflow:auto;">
    <div class="ui-widget-header">files</div>
    <div style="font-size:11px;" id="Menu">
<?php

$folders = folder_tree('*.xml', 0, $dirXML, -1);
//var_dump($folders);
print_r(renderTree($folders));
?>
    </div>
</div>    
<div class="page ui-state-default" style="height:100%;width:79%;right:0px;top:0px;bottom:0px;position:absolute;overflow:auto;">

    <label for="filename">Archivo:</label><input type="text" name="filename" value="<?php echo $dirXML2.$file; ?>" size="60" disabled="disabled"/>
<?php 
echo 'Perm: '.substr(sprintf('%o', fileperms($dirXML2.$file)), -4).' - ';

if (is_writable($dirXML2.$file)) {
    echo 'The file is writable';
} else {
    echo '<b>The file is not writable</b>';
}
?>
<button id="saveButton">Guardar Archivo</button>
<button id="indent">Identar</button>
<button id="undo">deshacer</button>
<button id="redo">rehacer</button>

<div id="ta" style="position:absolute; top:40px:bottom:2px;left:0px;right:0px;position:aboslute;">
<textarea id="textarea_1" name="content">
<?php


    if ($error != ''){
        echo $error;
    }
    else {
        $noutf = stripslashes(htmlentities($fileContent));
        $utf = stripslashes(htmlentities(utf8_decode($fileContent)));
        if ($utf != ''){
            echo $utf;
        } else {
            echo $noutf;
        }
    }
?>
</textarea>
</div>

    <span class="msg" style="visibility:hidden;" id="Msg" ></span>
</div>
    <script type="text/javascript">
    var textarea = document.getElementById('textarea_1');
    var myCodeMirror = CodeMirror.fromTextArea(textarea, {
    viewportMargin:Infinity,
    autoCloseTags: true,
    lineNumbers:true
  });

    $('#saveButton').click(function(){
	var content= myCodeMirror.getValue();
	saveFile(content);
	return false;
    }).button({icons: {primary:'ui-icon-disk'}});

    $('#indent').click(function(){
	var doc = myCodeMirror.getDoc();
	var lc = doc.lineCount();
	for(var ln = 0; ln <= lc ; ln++){
		myCodeMirror.indentLine(ln);
	    };
	
	return false;
    }).button({icons: {primary:'ui-icon-arrowstop-1-e'}});

    $('#undo').click(function(){
        var doc = myCodeMirror.getDoc();
        doc.undo();
    }).button({icons: {primary:'ui-icon-arrowreturn-1-w'}});

    $('#redo').click(function(){
        var doc = myCodeMirror.getDoc();
        doc.undo();
    }).button({icons: {primary:'ui-icon-arrowreturn-1-e'}});


/*
  var textarea = document.getElementById('textarea_1');
    editor = new MirrorFrame(CodeMirror.replace(textarea), {
    height: "auto",
    content: textarea.value,
    parserfile: "parsexml.js",
    stylesheet: "../lib/CodeMirror/css/xmlcolors.css",
    path: "../lib/codecirror/lib/",
    lineNumbers:true
  });
*/
    function saveFile (content) {
            //var content = $('#textarea_1').text();
            ShowMsg('Guardando Archivo: <?php echo $file; ?>');
            $("#Msg").load( "saveFile.php?file=<?php echo $file; ?>&dir=<?php echo $subdir; ?>", {contenido:content}, 
            function(){
            setTimeout(hideMsg, 1000);}
            );
            
    }


    function ShowMsg(msg){
            $('#Msg').css('visibility', 'visible').html('<img src="../img/Throbber-small.gif">' + msg);
    }

    function hideMsg(){
        $('#Msg').css('visibility', 'hidden')
    }
    $('li.boton').click(function(){
        $this = $(this);
       // alert($this.text());
       $dir = $this.parent('ul').attr('hrel');

       var destination = '?url=codeEditor.php&file=' + $this.text()+'&dir='+$dir;
        ShowMsg('Loading: '+ $this.text());       
       window.location= destination;
    });

</script>

</body>
</html>
<?php
function folder_tree($pattern = '*', $flags = 0, $path = false, $depth = 0, $level = 0) {
    $tree = array();
     
    $files = glob($path.$pattern, $flags);
    $paths = glob($path.'*', GLOB_ONLYDIR|GLOB_NOSORT);
     
    if (!empty($paths) && ($level < $depth || $depth == -1)) {
    $level++;
        foreach ($paths as $sub_path) {
	    $result = folder_tree($pattern, $flags, $sub_path.DIRECTORY_SEPARATOR, $depth, $level);
	
	    if (count($result) != 0)
                $tree[$sub_path] = $result;
        }
    }
 
    $tree = array_merge($tree, $files);
     
    return $tree;
}

function renderTree($tree, $dirname=''){
    global $actualFile;
    $salida = '<ul hrel="'.$dirname.'" >';
    foreach($tree as $filename => $file){
        if (is_array($file)){

            $dir = basename($filename);
            $salida .= '<li>'.$dir.'<br>'.renderTree($file, $dirname.'/'.$dir).'</li>';
        }
        else {
            $currentClass = '';                 

            $fileName = basename($file);
            if ($fileName == $actualFile)
                $currentClass = 'titulo';     
            
            if (!is_writable($file)){
                $currentClass .= 'gris';     
            }
	    if (is_file($file))
            $salida .= '<li class=" '.$currentClass.'">'.$fileName.'</li>';
        }
    }
    $salida .= '</ul>';
    return $salida;

}

?>