<?php
/* 
 * Histrix Data importer xml format
 */

include_once ("./autoload.php");
include_once ("../funciones/conexion.php");
include      ("./sessionCheck.php");

$filename = ($_GET['filename'] != '')? $_GET['filename'] : $filename;


// open zip file
if (is_file($filename)){
    $zip = new ZipArchive();
    $zip->open($filename);

    $Contxml = new Cont2XML();
    _begin_transaction();

     for($i = 0; $i < $zip->numFiles; $i++){
         
        if ($response === -1) continue;
        $innerFileName = $zip->getNameIndex($i);


        $fileContent   =  $zip->getFromName($innerFileName);
        $xmlData = simplexml_load_string($fileContent);
        $response = $Contxml->importData($xmlData);
     }
    _end_transaction();

    $zip->close();
    
    if ($response != -1) {
        $msg = '<span class="success">Import Success: '.$Contxml->inserts.'</span>';
    }
    else {
        $msg = '<span class="error">Import error</span>';
    }

    echo $msg.'<br>';
}


?>