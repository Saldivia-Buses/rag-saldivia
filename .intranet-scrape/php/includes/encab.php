<?php

if (!isset($_SESSION)) session_start();
if (!isset($path)) $path='';
if (!isset($inipath)) $inipath='';

if ($path != '' && $_GET['autoprint'] =='') {
// si lo llamo via ajax no cargo nada mas
// Salvo para la impresion
} else {
    if ($inipath != '')
        $path = $inipath;

    if (!isset($nom_empresa)) $nom_empresa = '';

?>
<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.1//EN" "http://www.w3.org/TR/xhtml11/DTD/xhtml11.dtd">
<html>
    <head>
        <title><?php echo htmlentities($nom_empresa, ENT_QUOTES, 'UTF-8');?></title>
        <meta name="apple-mobile-web-app-capable" content="yes" />
 
 
        <link rel="stylesheet" type="text/css"  href="<?php echo $path?>css/histrix.css" />


        <link rel="shortcut icon" href="<?php echo $path?>img/histrixico.gif" />

            <?php

            if ($client_is_mobile){
                echo '<meta name="viewport" content="width=320" />';

               echo '<link rel="stylesheet" type="text/css" href="'.$path.'css/mobile.css" />';

            }

            if (isset($_SESSION['css'])) {
                $cssFiles = $_SESSION['css'];
                if (is_array($cssFiles)) {
                    foreach($cssFiles as $n => $cssF) {
                        if(is_file($cssF)) {
                            $css_size = filesize($cssF);
                            echo '<link rel="stylesheet" type="text/css" href="'.$cssF.'?'.'" />';
                            echo "\n";
                        }
                    }
                }
                else {

                    if(is_file($path.'css/'.$_SESSION['css']))
                        echo '		<link rel="stylesheet" type="text/css" href="'.$path.'css/'.$_SESSION['css'].'" />';
                }
            }
            ?>
    </head>
    <?php
    }
?>