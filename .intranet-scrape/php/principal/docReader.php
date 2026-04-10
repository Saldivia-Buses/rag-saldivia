<?php
include ('./autoload.php');

$inipath = '../';
include ('../funciones/conexion.php');
include ("./sessionCheck.php");

?>
<body style="overflow:auto;">
    <?php

    $dirvar = $_GET['dir'];
    $file   = $_GET['f'];
    $tipo   = 'doc';


    $xmlPath = $datosbase->xmlPath;

    $basePath = '../database/'.$xmlPath.'/tmp/';

    $dirbase = $_SESSION[$dirvar];

    $url = $dirbase.'/'.$file;


    $url = $_GET['file'];

    $filesize  = @filesize  ($url);

    $filename  = md5_file($url);
    // converto to html
    $tmpdoc   = $basePath.$filename.'.doc';
    $tmphtml  = $basePath.$filename.'.html';
    
    if ($tipo == 'doc' ) {
            if (!is_file($tmphtml)) {

                copy($url, $tmpdoc);
                
                $command = 'abiword --to=html "'.$tmpdoc.'" ';
                echo passthru($command);
            }
    }
    else {
        copy($url, $tmppdf);
    }

    readfile($tmphtml);

    ?>
</body>
</html>
