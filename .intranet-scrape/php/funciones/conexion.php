<?php

// LIBRERIA PARA CONECTARSE A LA BASE DE DATOS
// Version 2.0 Objetos
//include_once ('config.php');
//include_once ('registry.php');

@include_once ('../config/config.php');
include_once (FUNCDIR . 'utiles.php');
include_once (FUNCDIR . 'func_odbc.php');
 // @include(LIBDIR . '/adodb/adodb.inc.php');

if (!isset($_SESSION)) {
    session_start();
}

$db = $_SESSION["db"];
$datosbase = Cache::getCache('datosbase'.$db);
if ($datosbase === false) {

    $xmlpath =( !isset($xmlpath) || ($xmlpath == '')  ) ? DBDIR : $xmlpath;
    $config = new config(CFGFILE, $xmlpath, $db);
    if (isset($config->bases[$db])) {
        $datosbase = $config->bases[$db];
        Cache::setCache('datosbase'.$db, $datosbase);
    }
}

if ($datosbase)
    $tipo_conex   = $datosbase->tipo;

// Include Lang Strings
if (is_object($datosbase)) {
    if ($datosbase->lang == '') $datosbase->lang = LANG;
    include_once(LANGDIR . $datosbase->lang . '.php');
}

if (isset($i18n))
    foreach ($i18n as $key => $val)
        $i18n[$key] = utf8_decode($val);



// Register objects
$registry =& Registry::getInstance();
$registry->set('lang', $datosbase->lang);
$registry->set('i18n', $i18n);
$dateFormat = 'd/m/Y';
$registry->set('dateFormat', $dateFormat);

$localxmlpath = (isset($datosbase->xmlPath))?$datosbase->xmlPath:'';
$dirXML = DBDIR . $localxmlpath . '/' . XMLDIR;
$registry->set('xmlPath', $dirXML );

$_SESSION['dirXML'] = $dirXML;


function getLink() {

    global $datosbase;
    global $tipo_conex;
    global $link;
    global $BASE;


    if (is_object($datosbase)) {
        // ME CONECTO
        $BASE = new Conexion($datosbase);
        // link es la conexion por defecto
        $link       = $BASE->link;
        //$linkMETA   = $BASE->linkMETA;

        $_SESSION['driver'] = $BASE->Driver;
        $_SESSION['tipo']   = $tipo_conex;
	return $link;
    }
    else {
         
    ?>
<html>
    <body>
        <div style="position:absolute;color:red;margin:center center center center;
             top:100px;font-size:12px;font-weight:700;border:1px solid orange;
             background-color:#ffee7f;padding:10px;">

		La Sesion se ha cerrado, cierre la ventana e ingrese al sistema nuevamente
        </div>
    <body>
</html>
    <?php
    }
}



/**
 *
 * @global <type> $link
 * @param <type> $sql
 * @param <type> $tipo
 * @param string $xml
 * @return <type>
 */
function updateSQL($sql, $tipo=null, $xml=null, $log=true) {

    global $link;

    if ($link=='') 
	$link = getLink();
    
    if ($log)
        loger($sql.';', 'updatessql.log', $xml);

    $result = _exec($link, $sql, $tipo);
    $error  = _error( $link);
    $errorNum = _errorNum($link);
    /*
    if (isset($_SESSION['EDITOR']) && $_SESSION['EDITOR'] == 'editor' ) {
        if (trim($xml) != '' && $sql != '')
            $_SESSION['last_update_sql'][$xml]= $sql;
    }
    */
    if ($error !='') {

      // TODO  move this code to this class
      //  $dbError  = new databaseError($error, $errorNum);

        $message  = 'Error de Update : '.$error."\n";

        $alerta .= 'Histrix.alerta("Error de Grabacion");';


        if(preg_match("/Duplicate entry/",$error, $matches)) {
            $customAlert = "Error,  Dato ya existente";
        }

        // MOVE THIS TO SEPARE CLASS
        if ($errorNum != '') {
            $pattern= '/REFERENCES \`([a-z0-9A-Z_]*?)\`/';
            preg_match($pattern , $error, $array);
            $referenceTable = $array[1];
            if ($referenceTable != '')
               $row = _fetch_array(consulta('SHOW TABLE STATUS WHERE Name = "'.$referenceTable.'"'));
            $nomTabla = explode(';',$row['Comment']);
        }
        
        if($errorNum == 1451) {
            $customAlert = sprintf(ERR1451,$nomTabla[0]);
        }

        if($errorNum == 1452) {
        	$customAlert = sprintf(ERR1452,$nomTabla[0]);
        }

        if ($customAlert != '') {
            $alerta = 'Histrix.alerta("'.$customAlert.'");';
        }
        //else
        {
        //    echo '<div class="error">('.$errorNum.') '. $error.'</div>';
        }

        // si hay Transaccion hago el rollback, sino, creo que no pasa nada.
        _rollback_transaction($message);

        $tableCreate = processError($errorNum, $error, $sql, $dir, $link);


        if ($alerta != ''){
            echo $error;
           // echo '<script type="text/javascript">'.$alerta.'</script>';
        }

        return -1;

        // DIE??
        //	die(''); don't die any more... let's see what happens
    }
    return $result;
}


/**
 * replace backslashes in SQL query
 * @param string $sql
 * @return string
 */
function replaceSlashes($sql) {
    $buscadas = array("\'" , '\"');
    $sustitutas =  array("''" , '"');
    $sqlfinal = str_replace ($buscadas, $sustitutas,$sql);
    return $sqlfinal;
}


function processError($errorNro, $errorMsg, $sql, $dir ='', $linkBase){

            if ($dir == ''){
                $registry = &Registry::getInstance();
                $dir = $registry->get('xmlPath');
            }
        loger('sql: '.$sql, 'search_schema.log');

            switch ($errorNro) {
                case 1146:
                    $regex = '/ \'(.*)\.(.*)\' /';
                    preg_match($regex, $errorMsg, $matches);
                    $database  = $matches[1];
                    $tableName = $matches[2];
                    $tableCreate = _loadSchemaFromFile($linkBase, $tableName, $dir );

                    break;
                case 1054: // unknown column
                    $regex = '/ \'(.*)\.(.*)\' /';
                    preg_match($regex, $errorMsg, $matches);
                    $tableName  = $matches[1];
                    $fieldName  = $matches[2];
                    
                    // IMPROVED ALTER TABLE RECOGNITION
                    if ($tableName == ''){
                        $regex = '/ into (.*) set /';
                        preg_match($regex, $sql, $matches);
                        $tableName  = $matches[1];
                        $regex = '/ column \'(.*)\' /';

                        preg_match($regex, $errorMsg, $matches);
                        $fieldName  = $matches[1];
                         
                    }
                    
                    // IMPROVED ALTER TABLE RECOGNITION
                    if ($tableName == ''){
                        $regex = '/ INTO (.*) SET /';
                        preg_match($regex, $sql, $matches);
                        $tableName  = $matches[1];


                        $regex = '/ column \'(.*)\' /';

                        preg_match($regex, $errorMsg, $matches);
                        $fieldName  = $matches[1];
    
                    }

                    if ($tableName ==''){
                        $regex = '/ FROM (.*) WHERE /';
                        preg_match($regex, $sql, $matches);
                        $tableName  = $matches[1];
                    }

                    $tableCreate = _loadSchemaFromFile($linkBase, $tableName, $dir , $fieldName);

                    break;

            }

            $htxError = new Histrix_Error($errorMsg);
            $htxError->send();

            return $tableCreate;
}

/**
 *
 * @global <type> $link
 * @global <type> $linkMETA
 * @global <type> $BASE
 * @global <type> $datosbase
 * @param <type> $consulta
 * @param <type> $enlace
 * @param string $opcion
 * @param string $xml
 * @return <type>
 */
function consulta($consulta, $enlace='', $opcion=null, $xml=null, $dir=null) {

    //$consulta = utf8_encode($consulta);  // check if this is optimall

    // LOGEO LAS CONSULTAS
    if ($opcion!='nolog') {

        if (isset($_SESSION['EDITOR']) && $_SESSION['EDITOR'] == 'editor' ) {
            loger($consulta.';', 'sql.log');

            /*
            if (trim($xml) != '' && $consulta != '')
                $_SESSION['last_select_sql'][$xml]= $consulta;
                */
        }

    }

    global $link;
    global $BASE;
    global $datosbase;

    if ($link=='') 
	$link = getLink();



    if (!$link) $link = $BASE->conectar($datosbase, SQL_CUR_USE_DRIVER);

    $linkBase = $link;
    $tableCreate = false;
    if ($link) {
        try {
            $result = _exec($linkBase, $consulta);
        } catch (Exception $ex) {
            print_r($ex);
        }
        if (!$result && $opcion != 'nofatal') {
        
            $errorNro = _errorNum($link);
            $errorMsg = _error($link);


            $tableCreate = processError($errorNro, $errorMsg, $consulta, $dir, $linkBase);


            if ($tableCreate) {
                $result = _exec($linkBase, $consulta);
                return $result;
            }

            $message  = 'Error de consulta.' . '<br>' . $errorNro . ' ' . $errorMsg . '<br>' . $consulta;

            loger($message, LOG_NAME_SELECT);

            if($enlace != 'firstrun')
            {
                if ($_SESSION['EDITOR'] == 'editor')
                    die('<div class="error">'.$message.'</div>');
                else
                    die('<div class="error">Error de Consulta.</div>');
            }
        }
    }
    else
    {
        $message = '<b>Error de Conexi&oacute;n</b>';

        if($enlace != 'firstrun')
            die($message );
        else
            echo($message );
       
    }

    return $result;
}

function campos($Tabla) {
    global $link;

    if ($link == '') {
	$link = getLink();
	}
    $result = _columns($link, '', '', $Tabla);

    return $result;
}

function getIndice($Tabla) {
    global $link;
    if ($link=='') 
	$link = getLink();

    $result = _primarykeys($link, '', '*', $Tabla);
    return $result;
}

function getCamposIndice($Tabla) {
//    global $linkMETA;
    global $link;
    global $BASE;
    if ($link=='') 
	$link = getLink();

    $result = _primarykeys($link, '', '*', $Tabla);

    while ($row = _fetch_array($result)) {
        if ($row['Key']=='PRI') {
            $campos[$row['Field']] = $row['Field'];
        }
    }
    return $campos;
}

/* CONVIERTO LOS FLOATS */

function mifloat($str, $set = FALSE) {
    if (preg_match("/([0-9\.,-]+)/", $str, $match)) {
        // Found number in $str, so set $str that number
        $str = $match[0];

        if (strstr($str, ',')) {
            // A comma exists, that makes it easy, cos we assume it separates the decimal part.
            $str = str_replace('.', '.', $str); // Erase thousand seps
            $str = str_replace(',', '', $str); // Convert , to . for floatval command

            return floatval($str);
        } else {
            // No comma exists, so we have to decide, how a single dot shall be treated
            if (preg_match("/^[0-9]*[\.]{1}[0-9-]+$/", $str) == TRUE && $set['single_dot_as_decimal'] == TRUE) {
                $str = str_replace('.', ',', $str);
                // Treat single dot as decimal separator
                return floatval($str);

            } else {
                // Else, treat all dots as thousand seps
                //   $str = str_replace('.', '', $str);    // Erase thousand seps
                return floatval($str);
            }
        }
    } else {
        // No number found, return zero
        return 0;
    }
}

function esClave($Tabla, $campo) {

    global $BASE;

    $rs = odbc_primarykeys2($Tabla, $campo);

    //	_result($rs, 'COLUMN_NAME');
    $cant = _num_rows($rs);
    for ($i = 0; $i < $cant; $i ++) {
        $clave = _result($rs, 'COLUMN_NAME');
        if ($clave == $campo)
            return true;
    }

    $array = _fetch_array($rs);
    if (($array))
        foreach ($array as $camp) {
            if ($array['COLUMN_NAME'] == $campo) {
                return true;
            }
        }
}

/* Funcion que simula la de PHP para ontener los campos clave de una tabla*/
function odbc_primarykeys2($Tabla, $campo = '') {
    global $link;
    if ($link=='') 
	$link = getLink();

    $result = odbc_primarykeys($link, '', '*', $Tabla);
    return $result;
}

/* Log Access into the database table */
function logAccess($login,$ip,$msg)
{
    updateSQL ( "insert into HTXACCESSLOG (login, ip, userAgent, error) values ('$login' , '$ip' , '{$_SERVER['HTTP_USER_AGENT']}' , '$msg' )", 'insert' );
}

?>