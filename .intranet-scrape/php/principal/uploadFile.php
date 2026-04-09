<?php

include_once ("./autoload.php");
include ("../funciones/conexion.php");
include ("./sessionCheck.php");

set_time_limit ( 200);

$db = $_SESSION["db"];
$datosbase = Cache::getCache('datosbase'.$db);
if ($datosbase === false) {
    $config = new config('config.xml', '../database/', $db);
    $datosbase = $config->bases[$db];
    Cache::setCache('datosbase'.$db, $datosbase);
}
//print_r( $_SERVER);

$base = dirname( $_SERVER['PHP_SELF']);
$inipath= $base.'/../';
include_once ('../includes/encab.php');

$nsf=$_FILES["file"]["name"];

$uploaddir = '../database/'.$_SESSION['datapath'].'/tmp/';


$path_info      = pathinfo($_FILES["file"]['name']);
$extension      = $path_info["extension"];
//$baseFileName   = $path_info["filename"];

$values    = $_REQUEST['values'];
$fields    = $_REQUEST['fields'];
$columns   = $_REQUEST['columns'];
$table     = $_REQUEST['table'];
$tipo	   = $_REQUEST['tipo'];
$firstLine = $_REQUEST['firstLine'];
$keyColumn = $_REQUEST['keyColumn'];
$keyField  = $_REQUEST['keyField'];
$postExec      = $_REQUEST['postExec'];

$style = 'display:none;';

if ($_REQUEST['debug'] == 1){
    $style= '';
}



function showConf($style) {

    global $values;

    global $fields;
    global $columns;
    global $table;
    global $tipo;
    global $firstLine;
    global $keyColumn;
    global $keyField;
    global $postExec;

    if ($fields != ''  || $keyField != '')  {


    ?>
    <table class="ui-widget-content" style="<?php echo $style; ?>">
        <tr><th>Tabla:</th><td><input type="text"  name="table" value="<?php echo $table;?>"></td></tr>
        <tr><th>tipo:</th><td> <input type="text"  name="tipo" value="<?php echo $tipo; ?>"></td></tr>
        <tr><th>Desde linea:</th><td><input type="text"  name="firstLine" value="<?php echo $firstLine; ?>"></td></tr>
            <?php
            if ($fields !='')
                foreach ($fields as $num => $nomField) {
                    
                    $columna = $columns[$num];
                    echo '<tr><td>columna: <input size="5" type="text"  name="columns['.$num.']" value="'.$columna.'"></td>';
                    echo '<td>Campo:<input size="10" type="text"  name="fields['.$num.']" value="'.$nomField.'"></td></tr>';
                }

            if ($values !='')
                foreach ($values as $num => $val) {
                    
    		    echo '<input size="5" type="hidden"  name="values['.$num.']" value="'.$val.'">';
                }


            if ($keyField !='')
                foreach ($keyField as $numk => $nomKField) {
                    echo '<tr><td>Clave: <input size="5" type="text"  name="keyColumn['.$numk.']" value="'.$keyColumn[$numk].'"></td>';
                    echo '<td>Campo:<input size="10" type="text"  name="keyField['.$numk.']" value="'.$nomKField.'"></td></tr>';
                }



            ?>
    </table>
<?php
    }
}


    ?>
<div class="Fichagrande" style="height:100%; overflow:auto">
<form action="uploadFile.php" method="post" enctype="multipart/form-data" >
    <p><?php 
    echo $_REQUEST['text'];
    ?></p>
    <fieldset>
        <legend>Importar Archivo </legend>

            <?php 
            
            showConf($style); 
            ?>
        <label for="file">Archivo:</label>
        <input type="file" name="file" id="file" />
        <input type="hidden" name="postExec" value="<?php echo $postExec; ?>" />
<?php
if ($nsf != ''){
    echo 'Archivo: ' . $nsf."<br/>";

    if (!in_array($extension, array('xls', 'csv', 'zip'))) {
        exit('Invalid extension.');
    } else if($_FILES["file"]["size"] > 2000000) {
            echo "File size too big!";
        } else {
            if ($_FILES["file"]["error"] > 0) {
                echo "Return Code: " . $_FILES["file"]["error"] . "<br />";
            } else {
                $pieces = explode(".", $_FILES["file"]["name"]);
                $importFilename = $pieces[0]."-".date("m.d.y").".".$extension;

                if (file_exists($uploaddir . $_FILES["file"]["name"])) {
                    echo $_FILES["file"]["name"] . " already exists.  Overwritting....<br/> ";
            }

            move_uploaded_file($_FILES["file"]["tmp_name"], $uploaddir.$importFilename );
          //  echo "Stored as: " . $uploaddir.$importFilename ;

            //echo 'extension: '. $extension;
            switch($extension){
            case 'csv':
                include 'csv_import.php';
            break;
            case 'xls':
                include 'xls_import.php';
            break;
            case 'zip':
                echo '<br/><label for="status">Status:</label>';
                $filename = $uploaddir.$importFilename;
                include 'importData.php';
            break;
            }

            if ($postExec != '') {
                 // post import Execution
                $command = '../cgi-bin/'.$postExec.'  2> /tmp/postExecError.log' ;

                loger($command, 'postExec.log');
                $sal = shell_exec($command);

                echo '<pre>'.$sal.'</pre>';

            }
        }
    }
}

?>
        <input type="submit" name="Enviar" value="Procesar" />
    </fieldset>
</form>
</div>