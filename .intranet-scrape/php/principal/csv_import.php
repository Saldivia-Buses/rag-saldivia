<?php
/*
 * Created on 11/08/2008
 * XLS DATA IMPORTER
 * needed data:
 INPUT FILE -->  DESTINATION TABLE
 //Insert
 ORIGEN DATA --> DESTINATION DATA
 //Update
 KEY DATA CONDITION ---> KEY COLUMNS
 *
 */

include_once ("./sessionCheck.php");
require_once '../lib/excel/lib/parsecsv.lib.php';


$csv = new parseCSV();

# Parse '_books.csv' using automatic delimiter detection...
$csv->auto($uploaddir.$importFilename);


$Datos = new ContDatos($table, 'IMPORTACION '.$table, $tipo);
$Datos->tipoInsert = 'IGNORE';
//_begin_transaction();

?>
 <h3>Datos Migrados</h3>
<div >
<table border="1" cellspacing="1" cellpadding="3">

<?php
// INSERT

	echo '<tr>';
 	foreach ($csv->titles as $value) {
		echo '<th>'.$value.'</th>';
 	}
 	echo '</tr>';

foreach ($csv->data as $key => $row) {

	$resultado = 'NO';
	$break = false;


	if ($columns !=''){
		echo '<tr>';
		foreach ($row as $col_name => $value){

			$index = array_search($col_name, $csv->titles) + 1; 
			$field_name = $fields[$index];

			$Datos->addCampo($field_name , '', '', '', $table, '');
			$field = $Datos->getCampo($field_name);
            $tipoDato = Types::getTypeXSD($field->TipoDato, 'xsd:string');

            if ($tipoDato == 'xsd:date'){

		$time = DateTime::createFromFormat('d/m/Y', $value)->format('Y-m-d');
            	/*$time = strtotime($value);
            	if ($time){
            		$value = date("Y-m-d", $time);
            	}*/
            }
            echo '<td>'.$value.'</td>';
			$Datos->setFieldValue($field_name , $value,'new');
			
		}
		echo '</tr>';
	}

	if ($values != ''){
	    foreach ($values as $col_name => $value){
			$field_name = $fields[$col_name];

			$Datos->addCampo($field_name, '', '', '', $table, '');
			$Datos->setFieldValue($field_name , $value, 'new');
	    }
	
	}

/*
	if ($keyColumn !=''){
		foreach ($keyColumn as $nkey => $key){
			$keyvalue = trim((string) $csv->sheets[0]['cells'][$i][$key]);
			if ($keyvalue =='') {
				$break = true;
			}
			$Datos->addCondicion($keyField[$nkey], '=', $keyvalue, ' and ', false);
		}
	}
	*/

	

	if (!$break){
		if ($tipo == 'insert'){

			$Datos->Insert();
		}
		else {
			if ($tipo == 'update'){
				//$Datos->getUpdate().'<br>';
				//$resultado = $Datos->getUpdate();

				$resultado =  $Datos->Update();


			}
		}
	}
	/*
	$info = _info(true);
	$matched = $info['rows_matched'];
	$changed = $info['changed'];
	echo '<td>'.$matched.'</td>';
	echo '<td>'.$info['changed'].'</td>';
	//echo '<td>'.$resultado.'</td>';
	echo '</tr>';
	
	$total['matched'] += $matched;
	$total['changed'] += $changed;
	*/
}

/*
foreach($total as $tipo => $tot){
	echo '<th>'.$tot.'</th>';
}
*/
//_end_transaction();

?>
</table>
</div>