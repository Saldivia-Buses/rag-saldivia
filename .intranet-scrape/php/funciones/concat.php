<?php
/*
 * Created on 17/11/2008
 *
 * histrix concat javascript librarie
 * thanks to http://verens.com/archives/2008/05/20/efficient-js-minification-using-php/
 * test
*/
include ("../principal/autoload.php");

if (!isset($_SESSION)) session_start();

$type= $_GET['type'];
$writabledir = '../'.$type.'/';

$datapath = $_SESSION['datapath'];

if ($datapath != '') $writabledir = '../database/'.$datapath.'tmp/';

$files = Cache::getCache($type);

switch($type) 
{
    case 'javascript':
        $ext = 'js';
        $header = 'Content-type: text/'.$type;
        break;
    case 'css':
        $ext = 'css';
        $header = 'Content-Type: text/'.$type.'; charset=ISO-8859-1';
        break;
}

$file_size = 0;
foreach($files as $n => $file) 
{
    if(is_file($file)) 
  	  $file_size .= filesize($file);
}

$newFileName = 'tmp'.md5($file_size).'.'.$ext;

if (!file_exists($writabledir.$newFileName)) 
{

    if (!is_dir($writabledir)) 
    {
        $createDir = mkdir($writabledir, 0777, true);
        if (!$createDir) {
            $error[]= 'Error Creating:'.$writabledir;
            $error[]= 'Check your '.$writabledir.' Permisions.';
            errorMsg($error);
            return  false;
        }
    }

	$filecontent = '';
    foreach($files as $n => $file) 
    {
  	
        if(is_file($file)) 
        {
      	    $fileContent .= "\n";
      	  	$fileContent .= '/* ***** BEGIN '.$file.' ******** */';
      	  	$fileContent .= "\n\n";
            $fileContent .= file_get_contents($file);
            $fileContent .= "\n";
                             
        }
    }
    // Delete old TMP
    deleteTmp($writabledir, $ext);
    file_put_contents($writabledir.$newFileName,$fileContent);
}


header($header);
header('Expires: '.gmdate("D, d M Y H:i:s", time() + 3600*24*365).' GMT');
readfile($writabledir.$newFileName);
















function deleteTmp($dir, $ext) {

    $files = scandir($dir);
    $midir = '';
    foreach($files as $nfile => $file) 
    {
        if (substr($file, 0, 3 ) == 'tmp' && substr($file, -3 ) == $ext) 
        {
            unlink($dir . $file);
        }
    }
    return;
}

