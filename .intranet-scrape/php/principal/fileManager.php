<?php
/*
 * Histrix File Manager
 * Luis M. Melgratti - 2008/09/08
 *
 */

include ("./autoload.php");
include ("./sessionCheck.php");

$inipath = '../';

$cssFiles[] =  '../css/histrix.css';
$cssFiles[] =  '../css/jquery.lightbox-0.5.css';
$cssFiles[] =  '../css/user.css.php';
$_SESSION['css']=$cssFiles;

include ('../includes/encab.php');
include ('../funciones/utiles.php');


$db = $_SESSION["db"];
$datosbase = Cache::getCache('datosbase'.$db);
if ($datosbase === false){
    $config = new config('config.xml', '../database/', $db);
    $datosbase = $config->bases[$db];
    Cache::setCache('datosbase'.$db, $datosbase);
}

?>
<body style="overflow:auto;">

<?php

$javascripts[] = 'sorttable.js';

//$_SESSION['javascript']= $javascripts;
Cache::setCache('javascript', $javascripts);

// Unified Javascript Libraries (less https calls)
echo '		<script type="text/javascript" src="../funciones/concat.php?type=javascript"></script>';
unset ($javascripts);

$javascripts[] = 'jquery-1.6.4.min.js';
$javascripts[] = 'jquery-ui-1.8.16.custom.min.js';

$javascripts[] = 'jquery.lightbox-0.5.min.js';
$javascripts[] = 'jquery.touch.compact.js'; // IPAD SUPPORT
$javascripts[] = 'jquery.histrix.filedrop.js';
$javascripts[] = 'jquery.idletimer.js';
$javascripts[] = 'histrix.js';
$javascripts[] = '../javascript/lang/histrix-'.$datosbase->lang.'.js';   

foreach($javascripts as $n => $js){
    $js_size = filesize('../javascript/'.$js);
    echo '<script type="text/javascript" src="../javascript/'.$js.'?'.$js_size.'"></script>';
    echo "\n";
}



$nom_empresa = $database->nombre;
$MarcaAgua   = $nom_empresa;
$img_fondo   = $database->img_fondo;
//$css   	     = $database->css;
$docPreview  = false;

$xmlPath=$datosbase->xmlPath;

$basePath = '../database/'.$xmlPath.'/xml/';

if (substr($basePath, -1 ,1) != '/')  $basePath .= '/'; // add trail slash

$basedir = trim($_REQUEST['basedir']);
$maxWidth = $_REQUEST['maxWidth']; //max image Width


if ($basedir != ''){
	$newdir = $basePath.$basedir;
	if (!is_dir($newdir))
    	@mkdir($newdir, 0777);
}

$dirvar = md5($basedir);

$home = trim($_REQUEST['home']);
if ($home != ''){
    $basePath = '../database/'.$xmlPath.'/files/';
	$basedir = 'home/'. $_SESSION['usuario'].'/';
	$newdir = $basePath.'home/'. $_SESSION['usuario'].'/';
    $_SESSION[$dirvar]['home'] 	 = 'true'; 		// current Dir
    if (!is_dir($newdir))
    	@mkdir($newdir, 0777, true);
    $access = 'rwd';
}

if (substr($basedir, 1 ,1) == '/')  $basedir = substr($basedir, 1 ); // remove trail slash


if ($_REQUEST != '')
foreach ($_REQUEST as $nreq => $req){
    ${$nreq} = $req;
    $_SESSION[$dirvar][$nreq] = $req;
}

$slashdir = $dir2;
/**
 * SESSION Variable sending
 */

if ($_REQUEST['dirvar']){
    $dirvar =  $_REQUEST['dirvar'];
    if ($_SESSION[$dirvar])
    foreach($_SESSION[$dirvar] as $nomvar => $var){
            ${$nomvar} = $var;
    }
}

if ($home != '')  $basePath = '../database/'.$xmlPath.'/files/';
if ($modo == '')  $modo = 'lista';
if ($access== '') $access = 'r';



// Sanitize Paths
$basePath = str_replace('//','/',$basePath);
$basedir  = str_replace('//','/',$basedir);
$slashdir = str_replace('//','/',$slashdir);
// Begins OOP implementation
$fileManager = new FileManager ($slashdir, $basePath, $access);
$fileManager->modo       = $modo;
$fileManager->inputField = $inputField;
$fileManager->dirvar     = $dirvar;
$fileManager->basedir    = $basedir;

$fileManager->docPreview = $docPreview;
$fileManager->maxWidth   = $maxWidth;
$salida .= '<div id="Supracontenido" ></div>';

$salida .= $fileManager->topBar($midir, $midir2, $access, $_FILES);

$salida .= Html::Tag('div', $fileManager->navPanel( $basePath.$basedir, $slashdir, 0) , array('id'=>'fileManager' ,
											      'destinationDir' => $basedir.$slashdir ,
											      'class' => 'fileManager TablaDatos',
											      'style' => 'z-index:10px',
											      'maxWidth' => $maxWidth
											      ));
$salida .='</div>';




echo $salida;

echo $fileManager->uploadJavascript();

//798
?>
<script language="JavaScript">
targetField = '<?php echo $inputField; ?>';
$('a[rel*=lightbox]').lightBox(); // Select all links that contains lightbox in the attribute rel
//sortables_init();
 $(document).ready(function(){
	            Histrix.init(window.opener, {confirmExit:false});
	            Histrix.db   = '<?php echo $_SESSION['db']; ?>';
	            Histrix.user = '<?php echo $_SESSION['usuario']; ?>';
				
		});

</script>
</body>
</html>
