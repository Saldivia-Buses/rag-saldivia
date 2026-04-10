<?php
/*
 * Created on 30/03/2006
 *
 * To change the template for this generated file go to
 * Window - Preferences - PHPeclipse - PHP - Code Templates
 */

class TablaDatos {
 	/* Array de Datos*/
//
//    var $Tabla;
//    var $MetaData;
//    var $nombre;
//    var $vacia;
//    var $indices;

    public function __construct($nombre='', $xml='') {
        $this->nombre = $nombre;
        $this->vacia = true;
    }

    /**
     * Insert a new record in the Temporal Table
     */
    public function insert($arrayin, $autoupdate= true, $replaceArray='') {
    // Si los campos tienen definidos indices se hace update de los registros y no solamente Insert
        $noup = false;


        if (isset($this->indices) && $this->indices != '' && $autoupdate === true) {
        // Busco registro

            $arraytemp = $this->indices;

            foreach($arraytemp as $campo => $valor) {
                $arraytemp[$campo] = $arrayin[$campo];
            }

            $hash = md5(implode($arraytemp));
            $row = $this->getRowUpdatebyHash($hash);

            //$row = $this->getRowUpdate($arraytemp);
            if (is_numeric($row)) {
                foreach($arrayin as $campin => $valin) {
                    if ($arraytemp[$campin] !='')
                        $this->Tabla[$row][$campin]= $valin;
                    else {
                        if (is_numeric ($valin ) ) {
                            
                            if (isset($replaceArray[$campin]))
                                $this->Tabla[$row][$campin]	= $valin;  // replace Value
                            else
                                $this->Tabla[$row][$campin]	+= $valin; // sum Value
                        }
                    }
                }
                $noup = true;
            }
            $arrayin['__HASH'] = $hash;

        }

        if (isset($arrayin['_ORDEN'])) {
            $arrayin['_ORDEN'] = count($this->Tabla) + 1;
        }

        if($noup == false) {
            $this->Tabla[]=$arrayin;
            $row = count($this->Tabla) ;
        }
        $this->vacia = false;

        return $row;
        
    }

    public function deleteRow($clavein) {
        if (isset($this->Tabla[$clavein]['_ORDEN'])) {
            $hayOrden = true;
        }
        unset($this->Tabla[$clavein]);
        unset($this->metaData[$clavein]);

		/* renumero */
        if ($hayOrden) {
            $num = 1;
            foreach($this->Tabla as $row => $fila) {
                $this->Tabla[$row]['_ORDEN'] = $num;
                $num++;
            }
        }
    }

	/* Limpia los campos Automaticos */
    public function deleteAutos() {
        if ($this->Tabla != '')
            foreach($this->Tabla as $orden => $row) {
                if ($row['_AUTO'] == true ) unset ($this->Tabla[$orden]);
            }
    }

    private function getRowUpdate($claves) {
        if ($this->Tabla != '')
            foreach($this->Tabla as $orden => $row) {
                $match = true;
                foreach($row as $nomcol => $valor) {
                    if($claves)
                        foreach($claves as $nomcla => $valcla) {
                            if ($nomcla == $nomcol&& $match != false)
                                if($valor == $valcla) $match = true;
                                else $match=false;
                        }
                }

                if ($match) return $orden;
            }
        return 'falso';
    }

    private function getRowUpdatebyHash($hash) {
        if ($this->Tabla != ''){
            foreach($this->Tabla as $orden => $row) {
                if ($row['__HASH'] == $hash )
                    return $orden;
            }
        }
    }

    private function getRow($claves) {
        foreach($this->Tabla as $orden => $row) {
            $borrar = false;
            foreach($row as $nomcol => $valor) {
                foreach($claves as $nomcla => $valcla) {
                    if ($nomcla == $nomcol)
                        if($valor == $valcla) $borrar = true;
                        else
                            return false;
                }
            }
            if ($borrar) return $orden;
        }
    }

    public function delete($claves) {
        $res = $this->getRow($claves);
        if ($res != false) $this->deleteRow($res);
    }


    public function emptyTable() {
        unset($this->Tabla);
    }


    public function update ($claves, $valores) {
        $res = $this->getRowUpdate($claves);
        $this->updateFila($res, $valores);
    }

    public function updateFila ($nrofila, $valores) {
    //	$resul = var_export($this->Tabla, true);

        $res = $nrofila;
        if (isset($this->Tabla[$res])) {

            foreach($this->Tabla[$res] as $columna => $valorcampo) {
                foreach($valores as $nomcolin => $nuevovalor) {
                    if ($columna == $nomcolin) {
                        $this->Tabla[$res][$columna] = $nuevovalor;
                    }
                    else {
                    //	loger($columna.'::'.$nomcolin, 'no');
                    }
                }
            }
        }
    }

    public function getMin($col) {
        $val =null;
        foreach($this->Tabla as $nfila => $fila) {
            $val = min($fila[$col], $val);
        }
        return $val;
    }

    public function getMax($col) {
        $val =null;
        foreach($this->Tabla as $nfila => $fila) {
            $val = max($fila[$col], $val);
        }
        return $val;
    }

    public function countRows() {
        return count($this->Tabla);
    }

    public function datos() {
        if (isset($this->Tabla))
            return $this->Tabla;
    }
    public function getRowData($order) {
        return $this->Tabla[$order];
    }

    public function setMetadata($row, $col, $data)
    {
        $this->metaData[$row][$col] = $data;
    }

    public function getMetadata($row, $col) 
    {
        if (isset($this->metaData[$row][$col])) {

            return $this->metaData[$row][$col];
        } else {
            //echo 'no hay';
        }
    }
    
    public function destroyMetadata($row, $col)
    {
        unset($this->metaData[$row][$col]);
    
    }



    public function newOrder($order){
        $orderArray = explode(',', $order);
        print_r($orderArray);
        $newTable = array();
        foreach( $orderArray as $i => $row){
            $newTable[$i] = $this->Tabla[$row];
        }
        $this->Tabla = $newTable;

    }

    public function ordenar($orden, $rev=false, $flags=0) {
        $orden = current($orden);
        if ($this->Tabla) {
            $sorted_records=$this->named_records_sort($this->Tabla, $orden, $rev, $flags);
        }
        $this->Tabla = $sorted_records;
    }

    function named_records_sort($named_recs, $order_by, $reverse=false, $flags=0) {
    // Create 1-dimensional named array with just
    // sortfield (in stead of record) values
        $named_hash = array();
        foreach($named_recs as $key=>$fields)
            $named_hash["$key"] = $fields[$order_by];

        // Order 1-dimensional array,
        // maintaining key-value relations
        if($reverse) arsort($named_hash,$flags=0) ;
        else asort($named_hash, $flags=0);

        // Create copy of named records array
        // in order of sortarray
        $sorted_records = array();
        foreach($named_hash as $key=>$val)
            $sorted_records["$key"]= $named_recs[$key];

        return $sorted_records;
    } // named_recs_sort()



    /**
     * Append another tempTable
     * @param TablaDatos $table
     */
    function appendTable(TablaDatos $table){
        $data = $table->Tabla;
        if (is_array($data))
        foreach($data as $row => $rowData){
            $this->insert($rowData);
        }
    }


}
 ?>