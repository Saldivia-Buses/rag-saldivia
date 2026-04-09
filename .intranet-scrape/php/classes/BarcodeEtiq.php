<?php
/*
 * Created on 04/07/2006
 * Luis M. Melgratti
 * Esta Clase Genera una etiqueta compatible con impresoras ZEBRA, lenguaje EPL2
 *
 */

class BarcodeEtiq{

	var $archivo;
	var $renglones;

	var $condicion;
	var $artcod;
	var $artdes;
	var $ctacod;
	var $ctanom;
	var $ocpnro;
	var $ocpfec;
	var $userid;
	var $codigounico;
	var $remfec;
	var $remnro;
	var $remnpv;


	function BarcodeEtiq($condicion = 1, $artcod='', $artdes='', $ctacod='', $ctanom='', $ocpnro='', $ocpfec='',$codigounico='', $remfec='', $remnro='', $remnpv='', $lote=''){

		$this->addEtiqueta($condicion, $artcod, $artdes, $ctacod, $ctanom, $ocpnro, $ocpfec,$codigounico, $remfec, $remnro, $remnpv, $lote);
	}

	function addEtiqueta($condicion, $artcod, $artdes, $ctacod, $ctanom, $ocpnro, $ocpfec,$codigounico, $remfec, $remnro, $remnpv, $lote=''){
		$opcion[1] = 'Aprobado';
		$opcion[2] = 'observado';
		$opcion[3] = 'Rechazado';
		$this->condicion = $opcion[$condicion];
		$this->artcod = $artcod;
		$this->artdes = substr($artdes, 0, 27);
		$this->ctacod = $ctacod;
		$this->ctanom = substr($ctanom, 0, 31);
		$this->ocpnro = $ocpnro;
		$this->ocpfec = $ocpfec;
		$this->codigounico = $codigounico;
		$this->remfec = $remfec;
		$this->remnro = $remnro;
		$this->remnpv = $remnpv;
		$this->lote = $lote;
		$this->userid = $_SESSION['usuario'];
	}

	// Seteo el Formato y tamaño de la etiqueta
	function setFormatoEtiqueta(){
		$this->addRenglon('O'."\n");
		$this->addRenglon('q790'."\n");
		$this->addRenglon('Q254,19'."\n");
	}

	/**
	* Blanueo Etiqueta
	*/
	function blanqueo(){
	    unset($this->renglones);
	}

	
	/**
	 * Genera el codigo de la Etiqueta
	 */
	function generaEtiqueta(){

		$y = 2;
		$this->addRenglon("\n");
		$this->addRenglon('N'."\n");
		$this->addRenglon('D10'."\n"); // Density 10
		$this->addRenglon('ZB'."\n");
		if ($this->setFormato != false){
			$this->setFormatoEtiqueta();
		}

		$this->addRenglon('A3,'.$y.',0,1,1,2,N,"Saldivia"'."\n");
        $this->addRenglon('A160,'.$y.',0,2,2,1,N,"'.$this->condicion.'"'."\n");
        $y += 23;
        $this->addRenglon('LO1,'.$y.',345,2'."\n");
        $y += 6;
        $this->addRenglon('A3,'.$y.',0,1,1,2,N,"('.$this->artcod.') '.$this->artdes.'"'."\n");
        $y += 26;
        $this->addRenglon('A3,'.$y.',0,1,1,2,N,"('.$this->ctacod.') '.$this->ctanom.'"'."\n");
        $y += 28;
        $this->addRenglon('A3,'.$y.',0,1,1,1,N,"Fecha: '.date('d/m/Y').'"'."\n");
        $y += 12;
        $this->addRenglon('A3,'.$y.',0,1,1,1,N,"O.Compra: '.$this->ocpnro.' Fec: '.$this->ocpfec.'"'."\n");

	$y += 12;
	$this->addRenglon('A3,'.$y.',0,1,1,1,N,"Remito  : '.$this->remnpv.'-'.$this->remnro.' Fec: '.$this->remfec.'"'."\n");
        $this->addRenglon('B4,149,0,1,2,7,50,B,"'.$this->codigounico.'"'."\n");

	$y = 2;
	$this->addRenglon('A428,'.$y.',0,2,2,1,R,"COPIA AL REMITO"'."\n");
        $y += 23;
        $this->addRenglon('LO426,'.$y.',345,2'."\n");
        $y += 6;
        $this->addRenglon('A428,'.$y.',0,1,1,2,N,"('.$this->artcod.') '.$this->artdes.'"'."\n");
        $y += 26;
        $this->addRenglon('A428,'.$y.',0,1,1,2,N,"('.$this->ctacod.') '.$this->ctanom.'"'."\n");
        $y += 28;
        $this->addRenglon('A428,'.$y.',0,1,1,1,N,"Fecha: '.date('d/m/Y').'"'."\n");
        $y += 12;
        $this->addRenglon('A428,'.$y.',0,1,1,1,N,"O.Compra: '.$this->ocpnro.' Fec: '.$this->ocpfec.'"'."\n");

	    $y += 12;
 		$this->addRenglon('A428,'.$y.',0,1,1,1,N,"Remito  : '.$this->remnpv.'-'.$this->remnro.' Fec: '.$this->remfec.'"'."\n");
        $this->addRenglon('B429,149,0,1,2,7,50,B,"'.$this->codigounico.'"'."\n");
        $this->addRenglon('P1'."\n");
        $this->addRenglon('N'."\n");

	}

	function addRenglon($renglon){
		$this->renglones[] = $renglon;
	}

	function Imprimir($printer='ZEBRA'){
		$fname = tempnam("/tmp", $this->codigounico.'eti');
		$fh=fopen($fname, "r+");
		loger($fname, 'etiq');
		loger($this->renglones, 'etiq');
		loger('end', 'etiq');
		
		foreach ($this->renglones as $nlinea => $linea){
			fwrite($fh, $linea);
		}
		exec('lp -d'.$printer.' '.$fname);
	}
}
?>