<?php
/* 
 * Loger Class 2013-04-28
 * Export class
 * @author Luis M. Melgratti
 */

class Export {
    
    function __construct(&$Container, $filename='') {
		$this->Container = $Container;	    	
    	$this->filename = $filename;

    	// exporting takes time
    	set_time_limit(300);
    }


    public function export(){
        $this->prepareFile();
        $this->removePagination();
        $this->header();
        $this->getData();
        $this->footer();
        $this->sendHeaders();
        $this->out();
    }

    public function removePagination(){
		// Reset Limit for Pagination cases
		if (isset($this->Container->paginar)){
		    // Redo the select statement
		    unset($this->Container->paginar);
		    unset($this->Container->limit);
		    $this->Container->Select();
		    $this->Container->CargoTablaTemporal();
		}
    }


    public function getData(){

        $Tablatemp = $this->Container->TablaTemporal->datos();
        $y=0;
        foreach ($Tablatemp as $orden => $row) {
            $y++;
            $x=0;
	    $fields = $this->Container->camposaMostrar();
	    //            foreach ($row as $nomcampo => $Valcampo) {
	    foreach($fields as $num => $nomcampo) {
		$Valcampo = $row[$nomcampo];
                if ($nomcampo =='') continue;

                $Field = $this->Container->getCampo($nomcampo);

                if ($Field->Oculto) continue;

                if (!is_object($Field) || isset($Field->export ) && $Field->export == 'false') continue;

                if ( $Field->Parametro['noshow'] == 'true') continue;

                $param['x'] = $x;
                $param['y'] = $y;
                $param['value'] = $Valcampo;
                $param['fieldname'] = $nomcampo;
                $this->outputData .= $this->processData($row, $Field, $param);

                $x++;
            }
            $this->endRow();
        }

    }

    public function prepareFile(){}
    
    public function processData($row, $field, $params){}

    public function endRow(){}

    public function header(){}

    public function footer(){}

    public function sendHeaders(){}

    public function out(){}


}
?>
