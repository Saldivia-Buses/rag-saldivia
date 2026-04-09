<?php
/*
 * Created on 11/05/2010
 *
 *  Temp Table
 */

class TempTable  extends TablaDatos {

    public function addRelationship( $id, $fieldName) {
        $this->relationship[$id] = $fieldName;
    }

    public function addRow( $row, $id, $fieldName) {
        $this->row[$row][$id] = $fieldName;
    }


    public function deleteData($xml){
	if(isset($this->Tabla)){
	    foreach($this->Tabla as $nrow => $row){
		if ($row['__xml__'] == $xml) unset($this->Tabla[$nrow]);
	    }
	}
    }


    /** fill data from another table with maping options from relationship
     *
     * @param TablaDatos $table
     */
    public function mapData(TablaDatos $tableObject, $xmlname) {
//        unset($this->Tabla); 		// delete current Data
	    $this->deleteData($xmlname); // delete JUST the currently calculated data, not imported data
        $data = $tableObject->Tabla;    // get Data from Table Object


        if (isset($data))
        foreach ($data as $order => $row) {

	    
            if (isset($this->row)) {
            
        	// for each row defined in the xml file it creates a new row and 
        	// insert it on the temptable
            
                foreach($this->row as $insertRow) {
                    foreach($insertRow as $id => $value) {

                        $arrayin[$id] = (isset($row[$value]))? $row[$value]: '';
                    }
                    $arrayin['__xml__'] = $xmlname;                    
                    
                    $this->insert($arrayin, false);
                    unset($arrayin);

                }

            }

	    // if there is a relationship
            if (isset($this->relationship)) {

                foreach($this->relationship as $target => $origin) {
                    $arrayin[$target] = (isset($row[$origin]))? $row[$origin] : '';;
                }
                $arrayin['__xml__'] = $xmlname;
               
                $this->insert($arrayin);
            }
        }
        
    }
}
 ?>