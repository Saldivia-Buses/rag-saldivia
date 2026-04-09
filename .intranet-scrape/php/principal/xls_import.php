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
require_once '../lib/excel/lib/reader.php';

// ExcelFile($filename, $encoding);
$data = new Spreadsheet_Excel_Reader();

// Set output Encoding.
$data->setOutputEncoding('CP1251');

//$importFilename = 'precios.xls';
$data->read($uploaddir.$importFilename);

//$table = $_GET['table']; //'LIBROS'
//$columns[]=1;
//$columns[]=2;
//$fields[]='ISBN';
//$fields[]='precioVenta';
//$tipo=$_GET['tipo']; 			//'update';
//$firstLine=$_GET['firstLine']; 	//2;
//PARA EL UPDATE
//$keyColumn[]=1;
//$keyField[] ='ISBN';

echo '<br>tipo: '.$tipo;


$Datos = new ContDatos($table, 'IMPORTACION '.$table, $tipo);
$Datos->tipoInsert = 'IGNORE';
_begin_transaction();

?>
 <h3>Datos Migrados</h3>
<div >
<table border="1" cellspacing="1" cellpadding="3">

<?php
// INSERT

	echo '<tr>';

 	for ($j = 1; $j <= $data->sheets[0]['numCols']; $j++) {
 		$tit = $data->sheets[0]['cells'][1][$j];
		echo '<th>'.$tit.'</th>';
 	}
 	echo '<th>Encontrado</th>';
 	echo '<th>Modificado</th>';
 	echo '</tr>';

for ($i = 1; $i <= $data->sheets[0]['numRows']; $i++) {

	if ($i < $firstLine) continue;
	$resultado = 'NO';
	$break = false;


	if ($columns !=''){
		foreach ($columns as $ncol => $col){

			$value = (string) $data->sheets[0]['cells'][$i][$col];

			$Datos->addCampo($fields[$ncol], '', '', '', $table, '');
			$Datos->setNuevoValorCampo($fields[$ncol], $value);

		}
	}

	if ($values != ''){
	    foreach ($values as $ncol => $value){
		$Datos->addCampo($fields[$ncol], '', '', '', $table, '');
		$Datos->setNuevoValorCampo($fields[$ncol], $value);

	    }
	
	}


	if ($keyColumn !=''){
		foreach ($keyColumn as $nkey => $key){
			$keyvalue = trim((string) $data->sheets[0]['cells'][$i][$key]);
			if ($keyvalue =='') {
				$break = true;
			}
			$Datos->addCondicion($keyField[$nkey], '=', $keyvalue, ' and ', false);
		}
	}
	echo '<tr>';
 	for ($j = 1; $j <= $data->sheets[0]['numCols']; $j++) {
		echo '<td>'.$data->sheets[0]['cells'][$i][$j];'</td>';
 	}
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
	$info = _info(true);
	$matched = $info['rows_matched'];
	$changed = $info['changed'];
	echo '<td>'.$matched.'</td>';
	echo '<td>'.$info['changed'].'</td>';
	//echo '<td>'.$resultado.'</td>';
	echo '</tr>';
	
	$total['matched'] += $matched;
	$total['changed'] += $changed;
}
// Totals
for ($j = 1; $j <= $data->sheets[0]['numCols']; $j++) {
	echo '<th></th>';
}

foreach($total as $tipo => $tot){
	echo '<th>'.$tot.'</th>';
}

_end_transaction();

?>
</table>
</div>