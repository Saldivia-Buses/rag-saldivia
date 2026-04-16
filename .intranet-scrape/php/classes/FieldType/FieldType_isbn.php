<?php
/* 
 * FieldType Class
 * 
 */

/**
 * Define FieldType representation
 *
 * @author luis
 */
class FieldType_isbn extends FieldType_varchar{

    const ALIGN   = 'left'; // Default Alignment
    const DIR     = 'ltr';  // Text direction
    const TYPE 	  = 'text';

    public function __construct(&$field=null){
        $this->field = $field;
    }

    public static function extraData(){
        return 'varchar';
    }

    public static function customValue($value, $field=null, $renderParameters=''){

		return $this->isbn102isbn13($value);
	}



	// ISBN13 a ISBN10
	private function isbn13_10($isbn) {
	    $ISBN = new ISBN($isbn);
	    return $ISBN->get_isbn10();
	}

	// ISBN10 a ISBN13
	private function isbn10_13($isbn) {
	    $ISBN = new ISBN($isbn);
	    return $ISBN->get_isbn13();
	}

	// OLD FUNCTION
	private function isbn102isbn13($b) {
	    return isbn10_13($b);
	}


}
?>
