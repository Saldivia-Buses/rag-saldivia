<?php
/** Clase usuario.
 * Almaceno las caracteristicas del usuario Actual
 * Nivel de acceso y propiedades
 */

class Histrix_Parameter extends Histrix{

    
    var $table;

    function __construct( $key, $label, $default, $login='' ) {
        parent::__construct();

        $this->key          = $key;
        $this->login        = $login;
        $this->label        = $label;
        $this->default      = $default;
        $this->table        = 'HTXOPTIONS';
        $this->key_field    = 'option_name';
        $this->login_field     = 'login';        
        $this->value_field  = 'option_value';
        $this->label_field  = 'option_description';        

        $this->getData();
    }

    public function getValue(){
        return $this->value;
    }
    /**
     * Get Data from Table
     */
    function getData() {
        $strSQL  = 'select '.$this->value_field.' from '.$this->table .' ';
        $strSQL .= 'where '.$this->key_field.' = "'. addslashes($this->key) . '"';
        if ($this->login != ''){
            $strSQL .= 'and '.$this->login_field.' = "'. addslashes($this->login) . '";';
        }

        $rs = consulta($strSQL);
        $i = 0;
        while ($row = _fetch_array($rs)) {
            foreach($row as $key => $value)
                $this->value = $value;
            $i++;
        }

        if ($i == 0){
            $this->insertData();
        }
    }

    function insertData(){
        $strSQL  = 'insert into '.$this->table .' ( '.$this->key_field.', '.$this->label_field.' , '.$this->value_field.', '.$this->login_field.')';
        $strSQL .= ' values ("'.$this->key.'", "'.$this->label.'" , "'.$this->default.'", "'.$this->login.'");';

        updateSQL($strSQL, 'insert');
        $this->value = $this->default;
    }

    function updateData($value){
        $strSQL = 'update '.$this->table.' set '.$this->value_field .'= "'.$value.'" where '.$this->key_field.'="'.$this->key.'";';
        updateSQL($strSQL, 'update');
    }


}

?>