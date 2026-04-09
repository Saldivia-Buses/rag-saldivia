<?php

/*
 * Created on 08/09/2006
 *
 * To change the template for this generated file go to
 * Window - Preferences - PHPeclipse - PHP - Code Templates
 */

 /**
  * Clase para hacer las minutas Contables
  */


class Minuta 
{
	/* estructura DE TODO ESTO (a poner en un xml) */

    var $CTBTBL 	= 'CTBCUENT';
    var $cuenta 	= 'ctbcod';
    var $ctbnombre 	= 'ctbnom';
    var $codsistema = 'siscod';
    var $ctbhab		= 'ctbhab';
    var $ctbcentro		= 'idcos';
    var $TBLNUMERADOR 	= 'CTBNUMER';
    var $campoNumerador = 'numapr';
    var $campoMes		= 'nummes';

    var $Plan; // Contenedor con el Plan de cuentas
    var $numero;
    var $ctbregnro;

    var $descripcion;
    var $fecha;
    var $referencia;
    var $totalDebe;
    var $totalHaber;

	// Array con renglones de las minutas
	var $renglones;

	var $valida;

	var $errores;

	public function Minuta($numero= '', $fecha = ''){
		$this->Plan = new ContDatos($this->CTBTBL, 'Cuentas','ayuda');
	$this->Plan->addCondicion($this->codsistema	, '=' , $this->getEjercicio(), 'and', 'reemplazo', false);

		if ($numero != '') $this->numero = $numero;
		if ($fecha != '') $this->fecha  = $fecha;

	}

 	public function Grabar(){

 		if ($this->validar()){
 			/* GRABO LA CABECERA DEL ASIENTO*/
 			// VER CUANDO SE IMPLEMENTE EL CTBMINCO

 			// Grabo movimientos

			$reng = $this->renglones;
			$i = 0;
			/* busco Numero */
			$this->numero = $this->GetNumero();

			foreach($reng as $nreng => $renglon){

				$i++;
 				$ContRenglon = new ContDatos($renglon->TBLminuta, 'Minuta', 'ayuda');

 				$ContRenglon->addCampo($this->codsistema , '', '', '', $renglon->TBLminuta, '');
 				$ContRenglon->addCampo('regmin' , '', '', '', $renglon->TBLminuta, '');
 				$ContRenglon->addCampo('regfec' , '', '', '', $renglon->TBLminuta, '');
 				$ContRenglon->addCampo('regref' , '', '', '', $renglon->TBLminuta, '');
 				$ContRenglon->addCampo('regnro' , '', '', '', $renglon->TBLminuta, '');
 				$ContRenglon->addCampo('idcos' , '', '', '', $renglon->TBLminuta, '');

 				$ContRenglon->addCampo('regord' , '', '', '', $renglon->TBLminuta, '');
 				$ContRenglon->addCampo($this->cuenta , '', '', '', $renglon->TBLminuta, '');
 				$ContRenglon->addCampo('regdoh' , '', '', '', $renglon->TBLminuta, '');
 				$ContRenglon->addCampo('regimp' , '', '', '', $renglon->TBLminuta, '');
 				$ContRenglon->addCampo('regpoa' , '', '', '', $renglon->TBLminuta, '');


 				$ContRenglon->setNuevoValorCampo($this->codsistema, $this->getEjercicio());
 				$ContRenglon->setNuevoValorCampo('regfec', $this->fecha);
 				$ContRenglon->setNuevoValorCampo('regref', $this->referencia);
 				$ContRenglon->setNuevoValorCampo('regnro', $this->ctbregnro);

 				$ContRenglon->setNuevoValorCampo('regord', $i);
 				$ContRenglon->setNuevoValorCampo($this->cuenta, $renglon->cuenta);
 				$ContRenglon->setNuevoValorCampo('regdoh', $renglon->doh);
 				$ContRenglon->setNuevoValorCampo('regimp', $renglon->importe);
 				$ContRenglon->setNuevoValorCampo('idcos', $renglon->ctbcentro_id);
 				$ContRenglon->setNuevoValorCampo('regpoa', 'A');


 				$ContRenglon->setNuevoValorCampo('regmin', $this->numero);


				echo $ContRenglon->Insert();
	 		}
	 		/* aumento el numerador */
	 		$this->AumentoNumero();

			return $this->numero;

 		}
 		else {
 			echo 'IMPOSIBLE GRABAR MINUTA ';
 			loger(print_r($this->errores, true), 'Minuta.log');
 			return -1;

 		}
 	}

 	/* ver el tema de manejar el ejercicio actual */
 	public function getEjercicio(){
		// ACA TENGO QUE IR A BUSCAR SEGUN EL EJERCICIO EN CURSO
 		return '22';

 	}

	/* obtengo el contenedor de numeracion actual */
	public function ContNumerador(){
 		$ejercicio = $this->getEjercicio();
		$ejercicio = '10';
		$tiempo = strtotime(Field::getValorGen($this->fecha, 'date'));
		$mes = date('m', $tiempo);

 		$ContNumerador =  new ContDatos($this->TBLNUMERADOR, 'Numerador', 'ayuda');

		$ContNumerador->addCampo($this->codsistema , '', '', '', $this->TBLNUMERADOR, '');
        $ContNumerador->setCampo($this->codsistema, $ejercicio);


		$ContNumerador->addCampo($this->campoNumerador , '', '', '', $this->TBLNUMERADOR, '');
		$ContNumerador->addCampo($this->campoMes , '', '', '', $this->TBLNUMERADOR, '');
		$ContNumerador->setOculto($this->campoMes, 'true');

		// condiciones
		$ContNumerador->addCondicion($this->codsistema	, '=' , $ejercicio, 'and', 'reemplazo', false);
		$ContNumerador->addCondicion($this->campoMes	, '=' , $mes, 'and', 'reemplazo', false);



		return 	$ContNumerador;
	}


 	/* ver el tema de manejar el Numero de Minuta Actual */
 	public function getNumero(){
		// Obtengo el contenedor del Numerador
		$ContNumerador = $this->ContNumerador();

		$ContNumerador->cargoCampos();
		$numerador = $ContNumerador->getCampo($this->campoNumerador)->getValor();

		return $numerador;
 	}

 	public function AumentoNumero(){
 		$ejercicio = $this->getEjercicio();
		$ejercicio = '10';
		$tiempo = strtotime(Field::getValorGen($this->fecha, 'date'));
		$mes = date('m', $tiempo);
		$ContNumerador = $this->ContNumerador();
                
		$ContNumerador->setCampo($this->codsistema	,  $ejercicio);
                $ContNumerador->setNuevoValorCampo($this->codsistema, $ejercicio);
		$ContNumerador->setCampo($this->campoMes	,  $mes);

		$proximoNumero = $this->numero + 1;
		$ContNumerador->getCampo($this->campoNumerador)->nuevovalor =  $proximoNumero;
 		$ContNumerador->Update();
 	}

 	/**
 	 * Anexo otra minuta a la actual
 	 */
 	public function anexoMinuta($minutaExterna){

 		if ($minutaExterna->renglones)
 		foreach($minutaExterna->renglones as $nreng => $renglon){
			if ($renglon->tipo != 'auto' && $renglon->cuenta != '')
			$this->addRenglon($renglon->cuenta, $renglon->doh, $renglon->importe, $renglon->tipo, $renglon->ctbcentro_id);
 		}
 	}
 	public function addRenglon($cuenta, $doh, $importe, $tipo='', $ctbcentro_id= ''){
//		$importe = round($importe, 2);
		if ($importe == 0) return false;
			$totdebe = 0;
			$tothaber = 0;

			if (($this->renglones ))
 			foreach($this->renglones as $nreng => $renglon){
 				if ($renglon->tipo == 'auto'){
	 				unset($this->renglones[$nreng]);
 					continue;
 				}
 			}

		$this->Plan->addCampo($this->cuenta		, '', '', '', $this->CTBTBL, '');
		$this->Plan->addCampo($this->ctbnombre	, '', '', '', $this->CTBTBL, '');
		$this->Plan->addCampo($this->ctbhab		, '', '', '', $this->CTBTBL, '');

		$this->Plan->addCondicion($this->cuenta	, '=' , "'".$cuenta."'", 'and', 'reemplazo', false);

 		$busqueda = $this->buscoRenglon($cuenta, $doh, $ctbcentro_id);
 		if ($busqueda != -1){
			$this->renglones[$busqueda]->importe += $importe;
			return;
 		}

 		$cantResultados = $this->Plan->cargoCampos();

		$nombre = $this->Plan->getCampo($this->ctbnombre)->getValor();

		$miRenglon = new RenglonMinuta($cuenta, $doh, $importe, $nombre, $tipo, $ctbcentro_id);
		// no existe la cuenta
		if ($cantResultados < 1) {
			$miRenglon->valido = false;
			$miRenglon->nombre = 'CUENTA INEXISTENTE: ' .$cuenta;
			$this->errores[3] = $miRenglon->nombre;

		}
		$hab = $this->Plan->getCampo($this->ctbhab)->getValor();

		if ($hab != 1) {


			$miRenglon->habil = false;
			$this->errores[4] = 'CUENTA '.$cuenta.'INHABILITADA';

		}
		$this->renglones[] = $miRenglon;

		foreach($this->renglones as $nreng => $renglon){
			if ($renglon->doh == 1)
				 $totdebe  += $renglon->importe;
			else $tothaber += $renglon->importe;
		}

		if ($tipo != 'auto'){
			$this->ajuste($totdebe, $tothaber);
		}

 	}

	// busco un renglon por cuenta y tipo, devuelvo el orden
 	public function buscoRenglon ($cuenta, $doh, $ctbcentro_id){
		$reng = $this->renglones;
		if (!(isset($reng))) return -1;
 		foreach($reng as $nreng => $renglon){
 			if ($cuenta == $renglon->cuenta && $doh == $renglon->doh && $ctbcentro_id == $renglon->ctbcentro_id)
 				return $nreng;
 		}
 		return -1;
 	}

 	// devuelve true o false si esta validado o no el asiento
 	public function validar(){

 		if ($this->fecha == ''){
 			$this->valida = false;
 			$this->errores[1]='MINUTA SIN FECHA';
 			return false;
 		}

 		/*if ($this->numero == ''){
 			$this->valida = false;
 			$this->errores[2]='MINUTA SIN NUMERO';
 			return false;
 		}*/

		$reng = $this->renglones;
		if (!(isset($reng))) {
 			$this->valida = false;
 			$this->errores[3]='MINUTA SIN MOVIMIENTOS';
 			return false;
		}

 		if (!($this->balanceado())) {
 			$this->valida = false;
 			$this->errores[4]= 'MINUTA DESBALANCEADA'. ' Dif: '.$this->dif;
 			return false;

 		}

		// chequeo caad renglon
 		foreach($reng as $nreg => $renglon){
 			if ($renglon->valido != true) {
 				$this->valida = false;
 				$this->errores[5]= 'CUENTA INEXISTENTE';
 				return false;
 			}

 			if ($renglon->habil != true) {
 				$this->valida = false;
 				$this->errores[6]= 'CUENTA INHABILITADA';
 				return false;
 			}
 		}


 		unset($this->errores);
 		return true;
 	}

	// Invierte la imputacion del asiento
 	public function invierto(){
 		foreach($this->renglones as $nreng => $renglon){
 			if ($renglon->doh == 1) $renglon->doh = 2;
 			else $renglon->doh = 1;
 		}
 	}

 	// devuelve true o false si esta balanceado o no el asiento
 	public function balanceado(){
		// Recorro el asiento y verifico si esta balanceado
		$reng = $this->renglones;
                $totdebe = 0;
                $tothaber = 0;
		if (isset($reng))
 		foreach($reng as $nreng => $renglon){
 			if ($renglon->doh == 1)
 				 $totdebe  += $renglon->importe;
			else $tothaber += $renglon->importe;
 //			echo 'renglon:'.$renglon->doh.$dif.'<br>';

 		}
		$dif = round($totdebe, 4) - round($tothaber, 4);
		$this->dif =$dif;
		//return $dif;
 		if ($dif== 0) return true;
 		else {
 			return false;
 		}

 	}

 	public function ajuste($totdebe, $tothaber){

		$ctb_debe  = $this->ajusteDebe;
		$ctb_haber = $this->ajusteHaber;
		$dif = $totdebe - $tothaber;
 		if ($dif < 0 ){
 			if ($ctb_debe != '') $this->addRenglon($ctb_debe, 1, abs($dif), 'auto');
 		}
 		else if ($ctb_haber != '') $this->addRenglon($ctb_haber, 2, abs($dif), 'auto');

 	}


	public function td($dat='', $opc=''){
		return '<td '.$opc.'>'.$dat.'</td>';
	}

	public function th($dat='', $opc=''){
		return '<th '.$opc.'>'.$dat.'</th>';
	}

	public function armovista(){
		$salida .= '<div class="TablaDatos"><table  width="100%" cellpadding="0" cellspacing="0" border="0" >';

		$salida .= '<tr>';
	 	$salida .= $this->th('Minuta Nro: '.$this->numero.'  '.$this->fecha, 'colspan="5"');
		$salida .= '</tr>';
		$salida .= '<tr>';
	 	$salida .= $this->th('Comprobante Nro'.$this->ctbregnro, 'colspan="5"');
		$salida .= '</tr>';

		$salida .= '<tr>';
	 //	$salida .= $this->th('Ref');
	 	$salida .= $this->th('Cuenta');
	 	$salida .= $this->th('Nombre');
	 	$salida .= $this->th('Debe');
	 	$salida .= $this->th('Haber');
	 	$salida .= $this->th('c.costos');

		$salida .= '</tr>';

		$reng = $this->renglones;
		if (isset($reng))
 		foreach($reng as $nreng => $renglon){
 			$salida .= '<tr>';
 			//$salida .= $this->td($this->referencia);
 			$salida .= $this->td($renglon->cuenta);
 			$salida .= $this->td($renglon->nombre);

 			if ($renglon->doh == 1){
 				 $totdebe  += $renglon->importe;
	 			$salida .= $this->td($renglon->importe);
	 			$salida .= $this->td('&nbsp;');
 			}
			else {
				$tothaber += $renglon->importe;
	 			$salida .= $this->td('&nbsp;');
	 			$salida .= $this->td($renglon->importe);
			}
			$salida .= $this->td($renglon->nombre_centro);

 			$salida .= '</tr>';
 		}
		$salida .= '<tr>';
	 	$salida .= $this->td('&nbsp;', 'colspan="2"');

	 	$salida .= $this->th($totdebe);
	 	$salida .= $this->th($tothaber);

		$salida .= '<tr>';
		if (isset($this->errores)){
			foreach ($this->errores as $nerr => $error){
				$saltemp.= $error;
			}
		}
	 	$salida .= $this->th($saltemp, 'colspan="4"');
		$salida .= '</tr>';

		$salida .= '<table></div>';
		return $salida;

	}

 	public function show(){
 		$this->validar();
 		return $this->armovista();
 	}

 }




?>