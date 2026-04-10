<?php

/*
 * Created on 03/12/2007
 *
 * To change the template for this generated file go to
 * Window - Preferences - PHPeclipse - PHP - Code Templates
 */

@include_once('../config/config.php');

if (!function_exists('loger')) {
function loger($string, $archivo = LOG_NAME_SYSTEM, $xml='') {
    $string = print_r($string, true);
    $fecha = date('d/m/Y H:i:s');
    $datedir = date('Y/m/');
    $xmlString = ($xml != '') ? ' (' . $xml . ')' : '';
    $string = $fecha . '|' . $_SESSION['usuario'] . $xmlString . ' - ' . $string . "\n";
    $datapath = $_SESSION["datapath"];

    $file = DBDIR . $datapath . LOGDIR . $datedir . $archivo;
    /*
      $recurso = fopen($file, 'a');
      fwrite($recurso, $string);
      fclose($recurso);
     */
    file_force_contents($file, $string, FILE_APPEND | LOCK_EX);
}
}

if (!function_exists('file_force_contents')) {

function file_force_contents($dir, $contents, $extra) {
    $parts = explode('/', $dir);
    $file = array_pop($parts);
    $dir = '';
    foreach ($parts as $part)
        if (!@is_dir($dir .= "/$part"))
            @mkdir($dir);

//            echo $contents;
  //  if (is_writeable("$dir/$file"))
        @file_put_contents("$dir/$file", $contents, $extra);
}
}
function errorMsg($error) {
    echo '<div class="error">';

    if (is_array($error))
        echo join($error, '<br />');
    else
        echo $error;
    echo '</div>';
}

/**
 * is_remote_access
 * Compare Server IP with User IP to determine if both are in same ip range
 *
 * @return boolean
 */
function is_remote_access() {

    $userIp  = array_slice(explode($_SERVER ['REMOTE_ADDR'], '.'), 0, 2);
    $localIp = array_slice(explode($_SERVER ['SERVER_ADDR'], '.'), 0, 2);

    return ($userIp == $localIp);
}

function is_utf8($str) {
    return (utf8_encode(utf8_decode($str)) == $str);
}

function pathinfo_utf($path) 
  { 
    if (strpos($path, '/') !== false) $basename = end(explode('/', $path)); 
    elseif (strpos($path, '\\') !== false) $basename = end(explode('\\', $path)); 
    else return false; 
    if (empty($basename)) return false; 

    $dirname = substr($path, 0, strlen($path) - strlen($basename) - 1); 

    if (strpos($basename, '.') !== false) 
    { 
      $extension = end(explode('.', $path)); 
      $filename = substr($basename, 0, strlen($basename) - strlen($extension) - 1); 
    } 
    else 
    { 
      $extension = ''; 
      $filename = $basename; 
    } 

    return array 
    ( 
      'dirname' => $dirname, 
      'basename' => $basename, 
      'extension' => $extension, 
      'filename' => $filename 
    ); 
  } 

function processing_time($START=false) {
    $an = 4;    // How much digit return after point
    if (!$START)
        return time() + microtime();
    $END = time() + microtime();
    return round($END - $START, $an);
}

function beginsWith($str, $sub) {
    return (substr($str, 0, strlen($sub)) == $sub);
}

function time_diff($horaini, $horafin) {

    $horai = substr($horaini, 0, 2);
    $mini = substr($horaini, 3, 2);
    $segi = substr($horaini, 6, 2);

    $horaf = substr($horafin, 0, 2);
    $minf = substr($horafin, 3, 2);
    $segf = substr($horafin, 6, 2);

    $ini = ((($horai * 60) * 60) + ($mini * 60) + $segi);
    $fin = ((($horaf * 60) * 60) + ($minf * 60) + $segf);

    $dif = $fin - $ini;

    $difh = floor($dif / 3600);
    $difm = floor(($dif - ($difh * 3600)) / 60);
    $difs = $dif - ($difm * 60) - ($difh * 3600);
    return date("H:i:s", mktime($difh, $difm, $difs));
}

function substrWords($str, $ini = 0, $end='') {
    $words = explode(' ', $str);
    if ($end == '')
        $end = count($words);
    for ($i = $ini; $i < $end; $i++) {
        $string .= $words[$i] . ' ';
    }
    return $string;
}

function NumeroALetras($tt) {
    $tt = $tt + 0.009;
    $Numero = intval($tt);
    $Decimales = $tt - Intval($tt);
    $Decimales = $Decimales * 100;
    $Decimales = Intval($Decimales);
    $salida = ucfirst(num2letras($Numero));

    If ($Decimales > 0) {
        $salida .= " con " . num2letras($Decimales) . " centavos";
    } else {
        $salida .= " con cero centavos";
    }
    return utf8_decode($salida);
}

/* !
  @function num2letras ()
  @abstract Dado un n?mero lo devuelve escrito.
  @param $num number - N?mero a convertir.
  @param $fem bool - Forma femenina (true) o no (false).
  @param $dec bool - Con decimales (true) o no (false).
  @result string - Devuelve el n?mero escrito en letra.

 */

function num2letras($num, $fem = false, $dec = true) {
//if (strlen($num) > 14) die("El n?mero introducido es demasiado grande");
    $matuni[2] = "dos";
    $matuni[3] = "tres";
    $matuni[4] = "cuatro";
    $matuni[5] = "cinco";
    $matuni[6] = "seis";
    $matuni[7] = "siete";
    $matuni[8] = "ocho";
    $matuni[9] = "nueve";
    $matuni[10] = "diez";
    $matuni[11] = "once";
    $matuni[12] = "doce";
    $matuni[13] = "trece";
    $matuni[14] = "catorce";
    $matuni[15] = "quince";
    $matuni[16] = "dieciseis";
    $matuni[17] = "diecisiete";
    $matuni[18] = "dieciocho";
    $matuni[19] = "diecinueve";
    $matuni[20] = "veinte";
    $matunisub[2] = "dos";
    $matunisub[3] = "tres";
    $matunisub[4] = "cuatro";
    $matunisub[5] = "quin";
    $matunisub[6] = "seis";
    $matunisub[7] = "sete";
    $matunisub[8] = "ocho";
    $matunisub[9] = "nove";

    $matdec[2] = "veint";
    $matdec[3] = "treinta";
    $matdec[4] = "cuarenta";
    $matdec[5] = "cincuenta";
    $matdec[6] = "sesenta";
    $matdec[7] = "setenta";
    $matdec[8] = "ochenta";
    $matdec[9] = "noventa";
    $matsub[3] = 'mill';
    $matsub[5] = 'bill';
    $matsub[7] = 'mill';
    $matsub[9] = 'trill';
    $matsub[11] = 'mill';
    $matsub[13] = 'bill';
    $matsub[15] = 'mill';
    $matmil[4] = 'millones';
    $matmil[6] = 'billones';
    $matmil[7] = 'de billones';
    $matmil[8] = 'millones de billones';
    $matmil[10] = 'trillones';
    $matmil[11] = 'de trillones';
    $matmil[12] = 'millones de trillones';
    $matmil[13] = 'de trillones';
    $matmil[14] = 'billones de trillones';
    $matmil[15] = 'de billones de trillones';
    $matmil[16] = 'millones de billones de trillones';

    $num = trim((string) @$num);
    if ($num[0] == '-') {
        $neg = 'menos ';
        $num = substr($num, 1);
    }else
        $neg = '';
    while ($num[0] == '0')
        $num = substr($num, 1);
    if ($num[0] < '1' or $num[0] > 9)
        $num = '0' . $num;
    $zeros = true;
    $punt = false;
    $ent = '';
    $fra = '';
    for ($c = 0; $c < strlen($num); $c++) {
        $n = $num[$c];
        if (!(strpos(".,'''", $n) === false)) {
            if ($punt)
                break;
            else {
                $punt = true;
                continue;
            }
        } elseif (!(strpos('0123456789', $n) === false)) {
            if ($punt) {
                if ($n != '0')
                    $zeros = false;
                $fra .= $n;
            }else
                $ent .= $n;
        }else
            break;
    }
    $ent = '     ' . $ent;
    if ($dec and $fra and !$zeros) {
        $fin = ' coma';
        for ($n = 0; $n < strlen($fra); $n++) {
            if (($s = $fra[$n]) == '0')
                $fin .= ' cero';
            elseif ($s == '1')
                $fin .= $fem ? ' una' : ' un';
            else
                $fin .= ' ' . $matuni[$s];
        }
    }else
        $fin = '';
    if ((int) $ent === 0)
        return 'Cero ' . $fin;
    $tex = '';
    $sub = 0;
    $mils = 0;
    $neutro = false;
    while (($num = substr($ent, -3)) != '   ') {
        $ent = substr($ent, 0, -3);
        if (++$sub < 3 and $fem) {
            $matuni[1] = 'una';
            $subcent = 'as';
        } else {
            $matuni[1] = $neutro ? 'un' : 'uno';
            $subcent = 'os';
        }
        $t = '';
        $n2 = substr($num, 1);
        if ($n2 == '00') {
            
        } elseif ($n2 < 21)
            $t = ' ' . $matuni[(int) $n2];
        elseif ($n2 < 30) {
            $n3 = $num[2];
            if ($n3 != 0)
                $t = 'i' . $matuni[$n3];
            $n2 = $num[1];
            $t = ' ' . $matdec[$n2] . $t;
        }else {
            $n3 = $num[2];
            if ($n3 != 0)
                $t = ' y ' . $matuni[$n3];
            $n2 = $num[1];
            $t = ' ' . $matdec[$n2] . $t;
        }
        $n = $num[0];
        if ($n == 1) {
            $t = ' ciento' . $t;
        } elseif ($n == 5) {
            $t = ' ' . $matunisub[$n] . 'ient' . $subcent . $t;
        } elseif ($n != 0) {
            $t = ' ' . $matunisub[$n] . 'cient' . $subcent . $t;
        }
        if ($sub == 1) {
            
        } elseif (!isset($matsub[$sub])) {
            if ($num == 1) {
                $t = ' mil';
            } elseif ($num > 1) {
                $t .= ' mil';
            }
        } elseif ($num == 1) {
            $t .= ' ' . $matsub[$sub] . 'ón';
        } elseif ($num > 1) {
            $t .= ' ' . $matsub[$sub] . 'ones';
        }
        if ($num == '000')
            $mils++;
        elseif ($mils != 0) {
            if (isset($matmil[$sub]))
                $t .= ' ' . $matmil[$sub];
            $mils = 0;
        }
        $neutro = true;
        $tex = $t . $tex;
    }
    $tex = $neg . substr($tex, 1) . $fin;
    return $tex;
}

function EscapeString($str) {
    $str = str_replace(array('\\', "'"), array("\\\\", "\\'"), $str);
    $str = preg_replace('#([\x00-\x1F])#e', '"\x" . sprintf("%02x", ord("\1"))', $str);

    return $str;
}

function strtolower_es($string) {

    $low = array("Á" => "á", "É" => "é", "Í" => "í", "Ó" => "ó", "Ú" => "ú", "Ü" => "ü", "Ñ" => "ñ");
    return strtolower(strtr($string, $low));
}

function strtoupper_es($string) {

    $low = array("á" => "Á", "é" => "É", "í" => "Í", "ó" => "Ó", "ú" => "Ú", "ü" => "Ü", "ñ" => "Ñ");
    return strtoupper(strtr($string, $low));
}

function resaltarStr($searchString, $value) {


    $pos = null;
    foreach ($searchString as $word) {

        if (strlen(trim($word)) > 2) {

            $posic = stripos($value, $word);
            if ($posic !== false)
                $pos = min($pos, $posic);
            $wordWidth = strlen($word);
            $substr = substr($value, $posic, $wordWidth);
            $replaceStr = '<i class="l_i_t">' . $substr . '</i>';
            $value = str_ireplace($word, $replaceStr, $value);

//            $value = preg_replace("/(".$word.")/i", '<i class="l_i_t">\0</i>', $value);
        }
    }
    return $value;
}

function hex_to_rgb($hex) {
//remievo el '#' si se encuentra
    if (substr($hex, 0, 1) == '#')
        $hex = substr($hex, 1);

    // si esta expresado en forma corta ('fff') lo expreso en forma grande
    if (strlen($hex) == 3) {
        $hex = substr($hex, 0, 1) . substr($hex, 0, 1) .
                substr($hex, 1, 1) . substr($hex, 1, 1) .
                substr($hex, 2, 1) . substr($hex, 2, 1);
    }

    if (strlen($hex) != 6) {
        // White default color
        $rgb['red'] = 255;
        $rgb['green'] = 255;
        $rgb['blue'] = 255;
    } else {
        //convierto de hexa a rgb
        $rgb['red'] = hexdec(substr($hex, 0, 2));
        $rgb['green'] = hexdec(substr($hex, 2, 2));
        $rgb['blue'] = hexdec(substr($hex, 4, 2));
    }
    return $rgb;
}


/**
 * Retrieve array of files of specified directory
 *
 * @param string $d Directory Path
 * @param string $x Files Filter
 * @return array
 */
function file_list($d, $x) {
    foreach (array_diff(scandir($d), array('.', '..')) as $f
        )if (is_file($d . '/' . $f) && (($x) ? ereg($x . '$', $f) : 1)
            )$l[] = $f;
    return $l;
}

/**
 *
 * @param string $datapath
 * @return boolean
 *
 */
function buildsystemstructure($datapath) {

    $path = DBDIR . $datapath;
    $firstrun = false;

    $result = createdir(DBDIR, $datapath);
    if ($result !== true) {
        errorMsg($result);
        return false;
    }

    $result = createdir($path, XMLDIR);
    if ($result !== true) {
        errorMsg($result);
        return false;
    }

    $result = createdir($path, TMPDIR);
    if ($result !== true) {
        errorMsg($result);
        return false;
    }

    $result = createdir($path, LOGDIR);
    if ($result !== true) {
        errorMsg($result);
        return false;
    }

    $result = createdir($path, IMGDIR);
    if ($result !== true) {
        errorMsg($result);
        return false;
    }

    $result = createdir($path, FILESDIR);
    if ($result !== true) {
        errorMsg($result);
        return false;
    }

    // Check Core Dir
    $dir = $path . XMLDIR . HTXDIR;
    if (!is_dir($dir)) {
        $firstrun = true;
    }

    // Check if TEMP folder got write access flags
    if( !is_writable($path . TMPDIR) ) {
        errorMsg($path . TMPDIR . ' folder write access error');
        return false;
    }

    // Verify Core Tables
    $result = @consulta('SELECT * from HTXUSERS LIMIT 0,1;', 'firstrun', 'nolog');
    if(!$result)
    {
        errorMsg('Some CORE Tables are Missing');
        return false;
    }

    return true;
}

/**
 *
 * @param string $path
 * @param string $name
 * @param string $symlink // If this parameter != null it will create a symlink to this
 *
 * @return mixed true if exist / error array if not exist
 *
 */
function createdir($path, $name, $symlink = null) {

    // TRIM LAST SLASH TO PREVENT ERROR

    if (substr($name, -1, 1) == '/')
        $name = substr($name, 0, -1);

    $dir = $path . $name;

    if (!is_dir($dir)) {

        if (!is_null($symlink))
            $createDir = @symlink($symlink, $dir);
        else {
            $createDir = @mkdir($dir, 0777, true);
        }

        if (!$createDir) {
            $error[] = 'Error Creating:' . $dir;
            $error[] = 'Check your [' . $path . '] permisions.';
            return $error;
        }
    }

    return true;
}

// Check if the file exists
// Check in subfolders too
function find_file ($dirname, $fname, &$file_path) {

  $dir = opendir($dirname);

  while ($file = readdir($dir)) {
    if (empty($file_path) && $file != '.' && $file != '..') {
      if (@is_dir($dirname.'/'.$file)) {
        find_file($dirname.'/'.$file, $fname, $file_path);
      }
      else {
        if (file_exists($dirname.'/'.$fname)) {
          $file_path = $dirname.'/'.$fname;
          return;
        }
      }
    }
  }

} // find_file


/**
 * Analyze a filenama and returns an array of path and filename
 */

function dirfile($str, $dir =''){
        $dirs['dir']  = $dir;
        $dirs['file'] = $str;
        // Force dir
        if (dirname($str) != '' && dirname($str) != '.'){
            $dirs['dir']  = dirname($str);
            $dirs['file'] = basename($str);
        }
       return $dirs;
}

?>
