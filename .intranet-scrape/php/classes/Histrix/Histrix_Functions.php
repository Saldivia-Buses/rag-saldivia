<?php

/**
 *
 * Functions
 *
 */
class Histrix_Functions {


    // TODO i18n
    const nameNegative   = "Menos";      // Precedencia para numeros negativos
    const nameCurrency   = " Peso";      // Nombre de Moneda en Singular
    const nameCurrencies = " Pesos";     // Nombre de Moneda en Plural
    const nameCent       = " Centavo";   // Nombre de Decimales en Singular
    const nameCents      = " Centavos";  // Nombre de Decimales en Plural
    const nameConcat     = "con";        // Preposición entre Moneda y Decimales

    static $isMoney      = false;
    
    /**
    *
    * Returns a money number into words 
    *
    * @param number         The desired number to be parsed
    * @param centsToString  When true cents will be parsed to words too
    * @param addCurrency    When false no currency name will be returned
    *
    ***************************************************************************/

    static function moneyToString($number, $centsToString = false, $addCurrency = true)
    {
      
      $result = '';
      
      self::$isMoney = $addCurrency;
      
  	  // Validates numbrer is under limits we support
  	  if ($number <= -1999999999.99 && $number >= 1999999999.99) return $result;	  

  	  // Get Money Part in words
  	  $result = self::numberToString( intval($number) );

  	  if ($addCurrency) 
  	      $result = $result . (abs($number) == 1 ? self::nameCurrency : self::nameCurrencies);

      // Get Cents in words
  	  $cents = intval( (abs($number)+0.009 - abs(intval($number))) * 100 );
    
      if( $cents > 0)
      {
        
        $result = $result . ' ' . self::nameConcat . ' ';
        $result = $result . ( $centsToString ? self::numberToString($cents) : "$cents/100" );
        
        if ($addCurrency) 
            $result = $result . ($cents == 1 ? self::nameCent : self::nameCents);
        
      }
      
      // This is Important bec it restore the default value of isMoney
      self::$isMoney = false;
	  	  
  	  return $result;
  	  
    
    }
    
    /**
    *
    * convert a number into words
    *
    *******************************************************/
    
    static function numberToString($number)
    {
    
      // TODO: i18n
	  $units    = array("", "Uno", "Dos", "Tres", "Cuatro", "Cinco", "Seis", "Siete", "Ocho", "Nueve", 
	                        "Diez", "Once", "Doce", "Trece", "Catorece", "Quince", "Dieciséis", "Diecisiete", "Dieciocho", "Diecinueve", "Veinte", 
	                        "Veintiuno", "Veintidos", "Veintitres", "Veinticuatro", "Veinticinco", "Veintiseis", "Veintisiete", "Veintiocho", "Veintinueve" );

	  $tens     = array("", "Diez", "Veinte", "Treinta", "Cuarenta", "Cincuenta", "Sesenta", "Setenta", "Ochenta", "Noventa", "Cien" );

      $hundreds = array("", "Ciento", "Doscientos", "Trescientos", "Cuatrocientos", "Quinientos", "Seiscientos", "Setecientos", "Ochocientos", "Novecientos" );
                        

      // Evaluates negative value and not zero
      $negative = ($number < 0) ? self::nameNegative.' ' : '';

      // ABS number
      $number = abs($number);
      
	  switch($number)
	  {
	  
		case 0:
		  $result = "Cero";
		  break;
		  
		case 1:
          $result = $units[$number];
          
      	  // Fix Un / Uno when no named money is converted
		  if ( self::$isMoney == true ) $result = 'Un';
		  break;
		
        case ($number <= 29):
          $result = $units[$number];
          break;
        
        case ($number <= 100):
          $result = $tens[ intval($number / 10) ] .      ($number %  10 != 0 ? ' y ' . self::numberToString($number %  10) : '');
          break;

        case ($number <= 999):
          $result = $hundreds[ intval($number / 100) ] . ($number % 100 != 0 ? ' '   . self::numberToString($number % 100) : '');
          break;



        case ($number <= 1999):
          $result = 'Mil' . ($number % 1000 != 0 ? ' ' . self::numberToString($number % 1000) : '');
          break;

        
        
    	case ($number <= 999999):
      	  $result = self::numberToString( intval($number / 1000) ) . ' Mil' . ($number % 1000 != 0 ? ' ' . self::numberToString($number % 1000) : '');
      	  break;

        case ($number <= 1999999):
      	  $result = "Un Millón" . ($number % 1000000 != 0 ? ' ' . self::numberToString($number % 1000000) : '');
      	  break;
      	  
      	case ($number <= 1999999999):
      	  $result = self::numberToString( intval($number / 1000000) ) . ' Millones' . ($number % 1000000 != 0 ? ' ' . self::numberToString($number % 1000000) : '');
      	  break;
	  }
	  
	  return $negative.$result;
  	  
    }


    public static function sessionClose(){
      include(dirname(__FILE__).'/../../config/config.php');
      // load Default Language
      $lang = substr($_SERVER['HTTP_ACCEPT_LANGUAGE'], 0, 2);
      $langfile = LANGDIR . $lang . '.php';

      if (file_exists($langfile)) {
          include_once($langfile);
          if (isset($i18n))
              $i18n  = array_map("utf8_encode", $i18n);
      }
      // 

      $message  = '<div  style="margin:20%;text-align:center; padding:10px;">';
      $message .= '<button style="box-shadow:2px 2px 15px #333;font-weight:700;font-size:12px; padding:10px;" onclick="window.location=\'..\'"><img style="vertical-align:middle" src="../img/emblem-important.png"/> <span> '.ERR1453.'</span></button>';
      $message .= '</div>';

//      $script[] = "Histrix.destroy('".ERR1453."'); ";
      $script[] = 'setTimeout(function(){  window.onbeforeunload = null; window.location="../";}, 10000);';
      $message .=  @Html::scriptTag($script);
      echo $message;

      session_start();
      session_unset();
      session_destroy();
      die();
        
    }

    /**
     * Remove unuses session info
     * @return [type] [description]
     */
    public static function garbageCollector()
    {

        $currenttime = time();
        $windows = $_SESSION['wintime'];

        foreach ($windows as $wid => $time) {
            //echo 'last time: '.$wid.' '.date('H:i:s',$time).'('. ( $currenttime - $time ) .')'."\n";

            // window has closed 300 secconds ago 
            // remove session information
            
            if ($currenttime - $time > 200) {
                $instances = $_SESSION['xml'];
                
                if ($wid != '') {
                    Histrix_Functions::removeInstances($instances, $wid);
                    unset($_SESSION['wintime'][$wid]);
                }
            }
        }

    }

    /**
     * Remove xml instaces stored in the currendt session in all or in some open windows
     * 
     * @param  array $instances list of instances
     * @param  string $window_id current window id
     * @return void            
     */
    public static function removeInstances($instances, $window_id='all')
    {
        
        foreach ($instances as $instance => $ContDatos) {
            //
            // Destroy Object
            // 
            $MisDatos = new ContDatos("");
            $MisDatos = Histrix_XmlReader::unserializeContainer(null, $instance);

            if ($window_id == 'all') {
                if (is_object($MisDatos) ) {
                    // destroy unused data;
                    $MisDatos->destroy();
                }

            } else {
                if (is_object($MisDatos) && $MisDatos->__winid == $window_id) {
                    // destroy unused data;
                    $MisDatos->destroy();
                }

            }
        }

    }
    
    /**
    * Test if internet connection is available
    */
    public static function hasInternetConnection(){
      
      if (!array_key_exists('internet', $_SESSION)){

        $fp = fsockopen("www.google.com", 80, $errno, $errstr, 30);
        if (!$fp) {
          $internet = false;
        } else {
          $internet = true;
        }
        $_SESSION['internet']  = $internet;



      }

	if ($fp)          
          fclose($fp);
      return $_SESSION['internet'];

    }

    
}
