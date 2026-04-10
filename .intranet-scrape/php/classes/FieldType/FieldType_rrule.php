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
class FieldType_rrule extends FieldType_varchar{

    const ALIGN = 'left'; // Default Alignment
    const DIR   = 'ltr';  // Text direction


    public function __construct(&$field=null){
    //    $this->field = $field;
    }
    
    public static function printValue($value){
    	return self::customValue($value);
    }

    public static function customValue($value, $field=null, $renderParameters=''){

		$rules = str_replace( 'RRULE:', '', $value);    
		$rules = explode(';' , $rules);
		$value = print_r($rules, true);
		
		$frequency['YEARLY'] = 'Anualmente';
		$frequency['MONTHLY'] = 'Mensualmente';
		$frequency['DAILY'] = 'Diariamente';
		$frequency['WEEKLY'] = 'Semanalmente';

		$freq2['YEARLY'] = 'Años';
		$freq2['MONTHLY'] = 'Mes/es';
		$freq2['DAILY'] = 'Días';
		$freq2['WEEKLY'] = 'Semana';

		$weekday['MO'] = 'Lunes';
		$weekday['TU'] = 'Martes';
		$weekday['WE'] = 'Miércoles';
		$weekday['TH'] = 'Jueves';
		$weekday['FR'] = 'Viernes';
		$weekday['SA'] = 'Sábado';
		$weekday['SU'] = 'Domingo';


		$months[1] = 'Enero';
		$months[2] = 'Ferebro';
		$months[3] = 'Marzo';
		$months[4] = 'Abril';
		$months[5] = 'Mayo';
		$months[6] = 'Junio';
		$months[7] = 'Julio';
		$months[8] = 'Agosto';
		$months[9] = 'Septiembre';
		$months[10] = 'Octubre';
		$months[11] = 'Noviembre';
		$months[12] = 'Diciembre';

		foreach($rules as $rule){

		    $values = explode('=',$rule);
		    $nameRule   = $values[0];
		    $valueRule  = $values[1];

		    switch ($nameRule) {
			case 'FREQ':
			    $freq  = $frequency[$valueRule];
			    $frequ2 = $freq2[$valueRule];
			break;
			case 'BYMONTH':
			    $month = $months[$valueRule];
			break;
			case 'BYMONTHDAY':
			    $day = $valueRule;
			break;
			case 'BYDAY':
			    $wday = $weekday[$valueRule];
			break;

			case 'INTERVAL':
			    if ($valueRule == 1) $valueRule = '';
			    $intervalo = 'cada '. $valueRule.' '.$frequ2;
			break;

		    }
		}
		
		$str = '';
		if ($freq != '')
		    $str  .= $freq.', ';

		if ($day != '')	    
		    $str .= ' cada '.$day;

		
		if ($month != '')
			$str .= ' de '.$month;
		
		if ($intervalo != ''){
		    $str = '';
			if ($day != '')	    
			    $str .= ' el dia '.$day.' de ';

		    $str .= $intervalo;

			if ($wday != '')	    
			    $str .= ' el dia '.$wday;
		    

		}

		if ($str != '')
	            $str .= '.';
		return $str;
		//. $value;
    }



}
?>
