<?php
/*
 * Created on 05/12/2007
 *
 * To change the template for this generated file go to
 * Window - Preferences - PHPeclipse - PHP - Code Templates
 */
 

 
 class Nodo{
 	var $data;
 	var $nodos = [];
	var $dataRow;
 	public function __construct($data){
 		$this->data = $data;
	
		
 	}
 	
 	public function addNodo($nodo){
 		$this->nodos[] = $nodo;
 	}
 	
    public function getNodes($key = '')
    {
        if ($key != '') {
            return $this->getNodesByKey($key);
        }
        return $this->nodos;
    }

    private function getNodesByKey($key)
    {
        $nodes = [];
        if ((string) $this->key == (string) $key) {
            return $this->nodos;
        }

        foreach ($this->nodos as $node) {
            $nodes = $node->getNodesByKey($key);
            if ($nodes != []) {
                return $nodes;
            }
        }
        return $nodes;
    }
 }
 
 
?>
