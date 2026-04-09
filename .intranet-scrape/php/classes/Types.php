<?php
/*
 * Created on 13/11/2007
 *
 * Clase estatica para manejo de Tipos
 */


class Types {
	
	public static function getTypeXSD($innerType, $default='xsd:string'){
		
		switch ($innerType){
			case 'tinyint':
				$xsdType = 'xsd:integer';
			break;
			case 'smallint':
				$xsdType = 'xsd:integer';
			break;
			case 'mediumint':
				$xsdType = 'xsd:integer';
			break;
			case 'integer':
				$xsdType = 'xsd:integer';
			break;
			case 'numeric':
				$xsdType = 'xsd:decimal';
			break;
			case 'decimal':
				$xsdType = 'xsd:decimal';
			break;
			case 'double':
				$xsdType = 'xsd:decimal';
			break;
			case 'float':
				$xsdType = 'xsd:decimal';
			break;
			case 'geoPoint':
                        case 'geoPoly':
			case 'editor':
			case 'simpleditor':
			case 'longtext':
			case 'mediumtext':
			case 'char':
			case 'email':
			case 'file':
			case 'month':
			case 'rrule':
			case 'varchar':
			case 'password':
				$xsdType = 'xsd:string';
			break;			
			case 'date':
				$xsdType = 'xsd:date';
			break;
			case 'time':
			case 'hora':
				$xsdType = 'xsd:time';
			break;
			case 'timestamp':
				$xsdType = 'xsd:dateTime';
			break;			
			case 'boolean':
				$xsdType = 'xsd:bolean';
			break;			
			default:
				$xsdType = $default;
			break;						
			
		}
		
		return $xsdType;
	}
	
        /**
        * Remove Quotes
        */
	public static function removeQuotes($value){
			// remove previous quotes if any
		if (substr($value, -1) == "'" && substr($value, 0,1) =="'"){
			$value =  substr($value, 1, strlen($value) - 2 );
		}
		
		if (substr($value, -1) =='"' && substr($value, 0,1) =='"'){
			$value =  substr($value, 1, strlen($value) - 2 );
		}
			

	    return $value;		
	}
	
	/**
	 * Get the value cuoted or not depending its type
	 */
	public static function getQuotedValue($value, $innerType, $defaultType){
		$xsdType= Types::getTypeXSD($innerType, $defaultType);
		if ($xsdType =='xsd:string' || $xsdType =='xsd:date' || $xsdType =='xsd:dateTime' || $xsdType =='xsd:time'){

			// remove previous quotes if any
			$value = Types::removeQuotes($value);
			// Addslashes prevents SQLiyection
			$value = addslashes($value);
			$newValue = "'".$value."'";
			if ($xsdType=='xsd:date' && $value == null) $newValue = $value;
		}
		else $newValue = $value;
		return $newValue;		
	}

    public static function formatDate($datetime, $dateFormat='d/m/Y'){
                
   		if (substr($datetime, 2, 1) != '/' && strlen($datetime) >= 8) {
			if ($datetime == '')
				return '';
			$tiempo = strtotime($datetime);

                        if (!$tiempo) return $datetime;
                        
			$date = date($dateFormat, $tiempo);
			return $date;
		}
		return $datetime;

    }

    public static function checkToday($value, $tipo){
		if ($value == 'today' && $tipo == 'date')
			$value = date("d/m/Y", time());
        return $value;
    }

}

?>