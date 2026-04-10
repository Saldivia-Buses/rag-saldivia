<?php
/*
 * Esta clase determina si se utilizan funciones ODBC o se redirigen a funciones mysql (u otro motor)
 * de forma explícita
 */

@include_once ('../config/config.php');

function _begin_transaction() {
    if ( $_SESSION['transaction'] == 'open') return;

    global $tipo_conex;
    global $BASE;
    global $link;
    if ($link=='') 
	$link = getLink();

    $link = $BASE->link;
    switch ($tipo_conex) {
        case "mysql":
            $link->autocommit(FALSE);
            //updateSQL('BEGIN;');
            break;
        case "ADODB":
            $link->BeginTrans();
            break;

    }
    loger('Begin Transaction', 'updatessql.log');
    $_SESSION['transaction'] = 'open';
}

function _end_transaction() {
    global $tipo_conex;
    global $BASE;
    global $link;
    if ($link=='') 
	$link = getLink();

    $link = $BASE->link;

    switch ($tipo_conex) {
        case "mysql":
            $link->commit();
            //updateSQL('COMMIT;');
            break;
        case "ADODB":
            $link->CommitTrans();
            break;
    }
    loger('commit', 'updatessql.log');
    $_SESSION['transaction'] = 'closed';
}

function _rollback_transaction($text='', $header=true) {
    global $tipo_conex;
    global $BASE;
    global $link;
    if ($link=='') 
	$link = getLink();
    
    $link = $BASE->link;
    switch ($tipo_conex) {
        case "mysql":
            $link->rollback();
            //updateSQL('ROLLBACK;');
            break;
        case "ADODB":
            $link->RollbackTrans();
            break;
    }
    loger('rollback '. $text, 'updatessql.log');
    $_SESSION['transaction'] = 'closed';

    if ($header)
    header('HTTP/1.1 400 Bad Request');


}

function _fetch_array($resource, $rownumber=null) {

    global $tipo_conex;
    switch ($tipo_conex) {
        case "mysql":
            if ($resource !='')
                return @$resource->fetch_array(MYSQLI_ASSOC);
            else return;
            break;
        case "ADODB":
            $ADODB_FETCH_MODE = ADODB_FETCH_ASSOC;
            //  $resource->SetFetchMode(ADODB_FETCH_ASSOC);
            //	if (is_object($resource)){
            //      $rs = $resource; //->GetAssoc();
            // }
          /*      echo '|'; */

            $rs = $resource->FetchRow();

            return $rs;
            //$rs = $resource->FetchRow();
            //else return;
            break;
        case "firebird":
            if ($resource !='')
                return ibase_fetch_assoc($resource, MYSQLI_ASSOC);
            else return;
            break;
    }
    return odbc_fetch_array($resource, $rownumber);
}

function _fetch_row($resource, $rownumber=null) {

    global $tipo_conex;
    switch ($tipo_conex) {
        case "mysql":
            return $resource->fetch_row();
            break;
        case "ADODB":
            $ADODB_FETCH_MODE = ADODB_FETCH_ASSOC;
            $rs = $resource->fetchRow();
            return $rs;
            break;
        case "firebird":
            return ibase_fetch_row($resource);
            break;
    }
    return odbc_fetch_row($resource, $rownumber);
}

/*
 function _result($result_id,  $field){

 	global $tipo_conex;
	switch ($tipo_conex){
		case "mysql":
			//if (is_numeric($field)) return mysql_result($result_id,  $field);

			// Ver si esto anda con mysql
			return mysqli_result($result_id,  $field);
		break;
	}
	return odbc_result($result_id,  $field);
 }
  */

/**
 * Devuelvo el ultimo ID del último Insert (Soportador por mysql solamente por ahora)
 */
function _insert_id($enlace=null) {

    global $tipo_conex;
    switch ($tipo_conex) {
        case "mysql":
            return $enlace->insert_id;
            break;
        case "ADODB":
            return $enlace->Insert_ID();
            break;
    }
    return 0;
}

function _connect($odbcUso, $user, $password, $cursor, $base, $host='localhost', $port = '3306') {
    global $tipo_conex;
    
    
    switch ($tipo_conex) {
        case "mysql":
        
            if (!class_exists('mysqli')) {
                die('<div class="error" >You need "mysqli" PHP extension.</div>');
            }

    	    $mysqli = mysqli_init(); 
    	    $mysqli->options(MYSQLI_CLIENT_COMPRESS,1);
    	    @$mysqli->real_connect( $host, $user, $password , $base, $port);

            if ($mysqli->connect_errno) {

                if ($mysqli->connect_errno == 1045) {
                    die('<div class="error">Invalid Database Configuration</div>');
                }

                die('<div class="error" >' . 'Connect Error: #' . $mysqli->connect_errno . '<br/>' . $mysqli->connect_error.'</div>');
            }
            

            while(true) {
            
            	$mysqli->select_db($base);

            	if ($mysqli->errno == 0) break;

	        	if ($mysqli->errno == 1049) {
                            $mysqli->query('create schema if not exists ' . $base);
                            continue;
	        	}
	        	
	        	die('<div class="error" >' . 'Connect Error: #' . $mysqli->errno . '<br/>' . $mysqli->error.'</div>');
	        }
	        			
            $mysqli->set_charset('utf8');

            //TODDO: to parametrize locale
            //ESTO ROMPE LA FACTURACION DE RESTIFFO ARREGLAR URGENTE
         //   $mysqli->query("SET lc_time_names = 'es_AR'");
          
            return $mysqli;
            break;
            
        case "ADODB":
            $DB = NewADOConnection('sybase');
            $DB->PConnect($host.':'.$base, $user, $password );
            return $DB;
            break;
        case "firebird":
            return ibase_connect($base, $user, $password);
            break;
        case "sybase":
            return sybase_connect($host, $base, $user, $password);
            break;

    }
    if($odbcUso)
        return odbc_connect($odbcUso, $user, $password, $cursor);
}

function _info($array='') {
    global $tipo_conex;
    global $BASE;
    global $link;
    if ($link=='') 
	$link = getLink();

    $link = $BASE->link;

    switch ($tipo_conex) {
        case "mysql":
            $info = $link->info;
            break;
        case "ADODB":
        //$info = $link->info;
            break;
    }

    if ($array == true) {
        $return = array();

        preg_match("/Records: ([0-9]*)/", $info, $records);
        preg_match("/Duplicates: ([0-9]*)/", $info, $dupes);
        preg_match("/Warnings: ([0-9]*)/", $info, $warnings);
        preg_match("/Deleted: ([0-9]*)/", $info, $deleted);
        preg_match("/Skipped: ([0-9]*)/", $info, $skipped);
        preg_match("/Rows matched: ([0-9]*)/", $info, $rows_matched);
        preg_match("/Changed: ([0-9]*)/", $info, $changed);

        $return['records'] = $records[1];
        $return['duplicates'] = $dupes[1];
        $return['warnings'] = $warnings[1];
        $return['deleted'] = $deleted[1];
        $return['skipped'] = $skipped[1];
        $return['rows_matched'] = $rows_matched[1];
        $return['changed'] = $changed[1];

        return $return;
    }

    return $info;
}

function _exec($connection_id, $query_string , $tipo=null) {

    global $tipo_conex;
    switch ($tipo_conex) {
        case "mysql":
            if ($tipo == 'insert') {
                $connection_id->multi_query($query_string);
            
		$autoinc = $connection_id->insert_id;

                while ($connection_id->more_results()) {
                    $connection_id->next_result();

                    } // flush multi_queries
                
                return $autoinc;
            }
            return $connection_id->query($query_string);

            break;

        case "ADODB":
            if ($tipo == 'insert') {
                $connection_id->Execute($query_string);
                return $connection_id->Insert_ID();

            //return $connection_id->insert_id;
            }

            return $connection_id->query($query_string);
            break;


        case "firebird":
            if ($tipo == 'insert') {
                ibase_query($connection_id, $query_string);
            //return mysql_insert_id();
            }
            return ibase_query($connection_id, $query_string);

            break;

    }

    return odbc_exec($connection_id, $query_string );
}

function _num_rows($result) {

    global $tipo_conex;
    switch ($tipo_conex) {
        case "mysql":
            return  $result->num_rows;
            break;
        case "ADODB":
            return  $result->RecordCount();
            break;
    }

    return odbc_num_rows($result);
}

function _num_fields($result) {

    global $tipo_conex;
    switch ($tipo_conex) {
        case "mysql":
        //   print_r($result);
            return @$result->field_count;
            break;
        case "ADODB":
        //   print_r($result);
            return $result->fieldCount();
            break;
    }

//	return odbc_num_fields($result);
}

function _field_name($result, $fieldnumber) {
    global $tipo_conex;
    switch ($tipo_conex) {
        case "mysql":
            $index = $fieldnumber - 1;
            $result->field_seek($index);
            $field = $result->fetch_field();
            return $field->name;
            break;
        case "ADODB":
            $index = $fieldnumber - 1;
            $field = $result->fetchField($index);
            return $field->name;
            break;
    }
    return odbc_field_name($result, $fieldnumber);
}

function _errorNum($connection_id=null) {

    global $tipo_conex;
    switch ($tipo_conex) {
        case "mysql":
            return $connection_id->errno;
            break;
        case "ADODB":
        // not implemented
        //return $connection_id->ErrorMsg();
            break;
        case "firebird":
        // not implemented
        //return ibase_errmsg( );
            break;
    }

    return odbc_error($connection_id);
}

function _error($connection_id=null) {

    global $tipo_conex;
    switch ($tipo_conex) {
        case "mysql":
            return $connection_id->error;
            break;
        case "ADODB":
            return $connection_id->ErrorMsg();
            break;
        case "firebird":
            return ibase_errmsg( );
            break;
    }

    return odbc_error($connection_id);
}


function _columns($connection_id , $qualifier='' , $schema='' , $table_name='' , $column_name='' ) {

    global $tipo_conex;
    
    switch ($tipo_conex) {
        case "mysql":
            return $connection_id->query('SHOW COLUMNS FROM '.$table_name.'' );
            break;
        case "ADODB":
            return $connection_id->Execute('SHOW COLUMNS FROM '.$table_name.'' );
            break;
    }

    return odbc_columns($connection_id , $qualifier , $schema , $table_name );

}

function _primarykeys($connection_id , $qualifier='' ,  $owner='' ,$table_name=''  ) {

    global $tipo_conex;
    switch ($tipo_conex) {
        case "mysql":
            return $connection_id->query('SHOW COLUMNS FROM '.$table_name.'' );
            break;
        case "ADODB":
            return $connection_id->Execute('SHOW COLUMNS FROM '.$table_name.'' );
            break;
    }

    return odbc_primarykeys($connection_id , $qualifier ,$owner, $table_name );
}

function _loadSchemaFromFile($link, $table, $dir='', $fieldName='') {

    $sqlFile  = $table.'.sql';
    $path     = DBDIR . $_SESSION['datapath'] . XMLDIR . $dir.'/'. SQLDIR;
    $fullFile = $path.$sqlFile;

    //if file is not found check in all subdirs
    if (!is_file($fullFile)){
        $newFile = '';
        loger('search: '.$sqlFile.' '.$fieldName, 'search_schema.log');
        

        find_file (DBDIR . $_SESSION['datapath'] . XMLDIR, $sqlFile, $newFile );
        $fullFile = $newFile;
        loger('found: '. $newFile, 'search_schema.log');

    }

    if (is_file($fullFile)) {

        $sql = file_get_contents($fullFile);
        $statements = explode(';', $sql);

        consulta('SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";',$link);

        // disable foreing keys
        consulta('SET FOREIGN_KEY_CHECKS = 0;',$link);

        foreach($statements as $num => $sqlStr) {

            if (strstr( $sqlStr, 'CREATE TABLE')) {

                if ($fieldName != '') {

                    $tableFields = campos($table);

                    //Lleno el Array de Campos del SQL (Necesario para obtener los tipos)
                    while ($fila = _fetch_array($tableFields)) {
                        $existingFields[$fila['Field']]= $fila;
                    }

                    //ADD FIELD
                    $ini= strpos($sqlStr,'(');
                    $strFields = substr($sqlStr, $ini + 1);
                    $fields= explode(",\n",$strFields);
                    foreach($fields as $nfield => $field) {
                        if (strpos(' '.trim($field), '`') === 1) {

                            $regex = '/ \`(.*)\` /';
                            preg_match($regex, $field, $matches);
                            $regFieldName = $matches[1];

                            if ($fieldName == $regFieldName)
                            if (!isset($existingFields[$regFieldName])) {
                                $addSel = "ALTER IGNORE TABLE $table ADD $field" ;
                                consulta($addSel,$link);
                                echo $addSel.'<br>';
                            }
                        }
                    }
                    return true;
                }
                else {
                // CREATE TABLE
                    $createTable = true;
                    consulta($sqlStr,$link);
                }
            }
            // re-enable foreing keys
            consulta('SET FOREIGN_KEY_CHECKS = 1;',$link);
            //
            if (strstr( $sqlStr, 'INSERT INTO') || strstr( $sqlStr, 'INSERT IGNORE INTO') && $createTable) {
                consulta($sqlStr,$link);
            }
        }
        return true;
    }
    else 
    {
	/*
	if ($_SESSION['EDITOR'] == 'editor'){
	    $sqlStr = 'CREATE TABLE IF NOT EXISTS `'.$table.'` (`'.$fieldName.'` as varchar) ENGINE=InnoDB  DEFAULT CHARSET=utf8;';
                              	    
	    echo $sqlStr;
                consulta($sqlStr,$link);
	
	}
	*/
        loger('Missing database table definition'.$fullFile);
    }
    return false;
}


?>
