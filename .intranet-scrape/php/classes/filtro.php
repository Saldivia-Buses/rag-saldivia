<?php
class filtro {

	/**
	* VER COMO UTILIZAR mejor esta clase, ahora solo se usa efectivamente si
	* el campo tiene un modificador en las condiciones (como un right)
	*/

	public $campo;
	public $operador;
	public $label;
	public $valor;
	public $opcion;
	public $deshabilitado;
	public $copia;	

	// Agrgado para premodificar el campo en el where
	public $modificador;

	/* Operadores posibles */

	public function __construct($campo, $oper, $label, $valor, $opcion='', $modificador = '', $grupo='', $deshabilitado=false, $copia='') {
		$this->campo = $campo;
		$this->operador = $oper;
		$this->label = $label;
		$this->valor = $valor;
		$this->opcion = $opcion;
		$this->modificador = $modificador;
		$this->grupo = $grupo;
		$this->deshabilitado = $deshabilitado;
		$this->copia = $copia;
		$this->uid = uniqid('F');								
	}
	
}

?>