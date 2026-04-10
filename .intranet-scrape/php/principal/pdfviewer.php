<?php
include ('./autoload.php');

$inipath = '../';
$DirectAccess=true;
include ('../funciones/conexion.php');
include ("./sessionCheck.php");

?>
<body style="overflow:auto;">
    <?php

    $javascripts[] = 'swfobject.js';

    foreach($javascripts as $n => $js) {
        $js_size = filesize('../javascript/'.$js);
        echo '<script type="text/javascript" src="../javascript/'.$js.'?'.$js_size.'"></script>';
        echo "\n";
    }

    $dirvar = $_GET['dir'];
    $file   = $_GET['f'];
    $tipo   = $_GET['tipo'];
    $readonly   = $_GET['ro'];

    $database = $_SESSION['datosbase'];

    $db = $_SESSION["db"];
    $database = Cache::getCache('datosbase'.$db);
    if ($database === false) {
        if ($xmlpath == '')  $xmlpath = '../database/';
        $config = new config('config.xml', $xmlpath, $db);
        $database = $config->bases[$db];
        Cache::setCache('datosbase'.$db, $database);
    }

    $xmlPath = $datosbase->xmlPath;

    $basePath = '../database/'.$xmlPath.'/tmp/';


    $dirbase = $_SESSION[$dirvar];

    if ($dirvar == ''){
        $dirbase = '../files/';
    }

    if (file_exists($basePath.$dirbase))
        $dirbase = $basePath.$dirbase;


    $url = urldecode($dirbase.'/'.$file);


    if (!is_file($url)){
	echo 'no encontado: '.$url;
    
    }
    else {
    }
    $filesize  = @filesize  ($url);
    $filename  = md5_file($url);


    // if type is not set
    if ($tipo == ''){
        $tipo = substr(basename($url), strrpos(basename($url), ".")+1);
        if (is_file($url)) {
            $path_info 		= pathinfo($url);
            $extension 		= $path_info["extension"];
            $tipo = $extension;
        }
    }


    $tmpswf  = $basePath.$filename.'.swf';
    $tmppdf  = $basePath.$filename.'.pdf';
    $tmppdf2 = $basePath.$filename.'.pdf';
    $tmpsvg  = $basePath.$filename.'.svg';
    $pdf = true;
    $zoom = 80;
    if ($tipo == 'txt') {
        echo '<pre>';
        echo readfile($url);
        echo '<pre>';
        $pdf = false;
    }
    else
    if ($tipo == 'dwg') {
        // CONVERTIR SVG A JPG
        $filenamesvg = uniqid('svg');
        $tmpdwg0 = $basePath.$filename.'.dwg';
        copy($url, $tmpdwg0);
        if (!is_file($tmpsvg)) {
            $command = '../cgi-bin/cad2svg  "'.$tmpdwg0.'"  -o '.$tmpsvg;
            echo $tmpsvg;
            exec($command);
        }
        unlink ($tmpdwg0);

        // FIX BLACK BACKGROUND
        $command = '../cgi-bin/htx_svg_fix  "'.$tmpsvg.'" '.$tmppdf;
        exec($command);
        //echo $command;

        if (!is_file($tmppdf)) {
            $command = '/usr/bin/convert '.$tmpsvg.' '.$tmppdf;
            exec($command);
        }
        $zoom=40;

    }
    else {
        if ($tipo == 'doc' ||
                $tipo == 'xls' ||
                $tipo == 'odt' ||
                $tipo == 'ods' ||
                $tipo == 'odt' ||
                $tipo == 'odp'  ) {

            $uniqid = $basePath.$filename.'.'.$tipo;

            // requires antiword application
            //$command = '/usr/bin/antiword -m 8859-1 -a a4 "'.$url.'" > '.$tmppdf;

            // oooconv converter (SLOW BUT BETTER)
            if (!is_file($tmppdf)) {

                copy($url, $uniqid);

                // oooconv converter (SLOW BUT BETTER)
                // $command = '../cgi-bin/oooconv "'.$uniqid.'"  '.$tmppdf. ' > /tmp/convlog.log';
                // oo2pdf Script best for now..
                //$command = '../cgi-bin/oo2pdf '.$uniqid.'  ';
                $command = 'abiword --to=pdf "'.$uniqid.'" ';

                echo passthru($command);
             //   echo $command;


            }


        }
        else {
                copy($url, $tmppdf);
        }
    }
    if ($pdf) {
            if (is_file($tmppdf))
                copy($tmppdf, $tmppdf2);

            $viewport = 'fdviewer.swf';


            if (!is_file($tmpswf)) {

                
//                $command1 = 'convert '.$tmppdf.'  '.$tmppdf;
//                exec($command1);
                $command2 = 'pdf2swf  -s zoom='.$zoom.' -o '.$tmpswf.' '.$tmppdf;
                exec($command2);
            }

            $filepath = $basePath.'V'.$filename.'.swf';
            //$viewport = 'SimpleViewer.swf';


            if (!is_file($filepath)) {
                $command3 = 'swfcombine  -X 1024 -Y 768W -o '.$filepath.' '.$viewport.' \'#1\'='.$basePath.$filename.'.swf ';
                exec($command3);
            }


            $archivo 	= new Archivo($basePath.'V'.$filename.'.swf', $basePath, '');


            echo $archivo->swfViewer($basePath.'V'.$filename.'.swf');

            //$nombre = utf8_decode($file);
            $nombre = basename($file);

            if ($readonly == ''){
                echo $archivo->downloadButton($url, $nombre);
                echo $archivo->printButton($tmppdf2, $nombre);
            }
        

    }
    ?>
</body>
</html>
