<?php
/**
 * Clase que almacena las posibles condiciones de un campo para una sentencia SQL
 */
class Condicion {
    var $label;
    var $OpLogico;
    var $operador;
    var $valor;
    var $premodificador;
    var $verdadero;
    var $falso;
    var $tipo;
    var $grupo;
    var $fixed;

	/* Operadores posibles */

    function __construct($Logico, $oper, $val, $label = '', $premodificador='', $grupo=0, $fixed=false) {

        $this->OpLogico = $Logico;
        $this->operador = $oper;
        //echo utf8_decode($val);
        $this->valor = $val;
        $this->label = $label;
        $this->premodificador = $premodificador;
        $this->grupo = $grupo;
        $this->fixed = $fixed;

    }

    /**
     * Devuelve la condicion logica
     */
    public function armaCondicion($campo, $Pri) {
        $valor = $this->valor;
        switch ($this->premodificador ) {
            case 'right' :
                $sincomillas = trim($valor, '"');
                $sincomillas = trim($valor, "'");
                if (strtoupper($this->operador) == 'LIKE'){
                    $valor = "'%".$sincomillas."'";
                } else {
                    $campo = 'right(cast('.$campo.' AS CHAR), '.strlen($sincomillas).' ) ';
                }
                break;
            case 'left' :
                $sincomillas = trim($valor, '"');
                $sincomillas = trim($valor, "'");
               if (strtoupper($this->operador) == 'LIKE'){
                    $valor = "'".$sincomillas."%'";
                } else {
                    $campo = 'left(cast('.$campo.' AS CHAR), '.strlen($sincomillas).' ) ';
                }
                break;
        }
        $OpLogico = '';

        if (!$Pri) {
            $OpLogico = $this->OpLogico;
        }
        if (is_array($valor)){
            $valor = implode(',', $valor);
            $this->operador  = 'in';
        }
        

        switch ($this->operador) {
            case 'in':
        	$valor = trim($valor);
		$valor = ltrim($valor,'(');
		$valor = rtrim($valor,')');
		
                $cadena = $OpLogico.' '.$campo.' '.$this->operador.' ('.$valor.') ';
                break;
            case 'isnull':
                $cadena = $OpLogico.'  isnull('.$campo.')';
            break;
            
            default:
                $cadena = $OpLogico.' '.$campo.' '.$this->operador.' '.$valor.' ';                    
            break;
        }

        //echo $cadena;
        if (strlen($valor) >= 1) {

            return $cadena;
        }
    }
}

?>