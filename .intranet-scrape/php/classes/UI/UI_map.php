<?php

/*
 * 2011-10-25
 * help popup class
 */
//include "../lib/RRule/RRule.php";
//include "../lib/When/When.php";

class UI_map extends UI_consulta {

    /**
     * User Interfase constructor
     *
     */
    public function __construct(&$Datacontainer) {
        parent::__construct($Datacontainer);
        $this->tag = 'div';
        $this->hasForm = true;
        $this->defaultClass = 'consulta';

        $this->javascriptLibs[] = 'https://maps.google.com/maps/api/js?libraries=places&sensor=false&callback=initGmap_'.$this->Datos->idxml;
    //    $this->sidePanel = (isset($this->Datos->sidePanel) && $this->Datos->sidePanel != '') ? $this->Datos->sidePanel : true;
    }

    private function googleMapsInit() {



        /*
         eventSources: [ 'process.php?reload=true&xmldatos=".$this->Datos->xml."&instance=". $this->Datos->getInstance()."' ]

         */
        $iniLat = '-32.945616471035706';
        $iniLon = '-60.63649535179138';

        $iniLat = '-33.03062';
        $iniLon = '-60.61451';

        $jscode= "

        var initGmap_".$this->Datos->idxml."= function(){

            var latlng = new google.maps.LatLng($iniLat, $iniLon);
            var my_map = $('#".$this->Datos->idxml."');
            my_map.css({width:'auto', height:'100%'});
        
            var myOptions = {
              zoom: 15,
              center: latlng,
              mapTypeId: google.maps.MapTypeId.ROADMAP
            };
        
            Histrix.gmaps['".$this->Datos->idxml."'] = new google.maps.Map(my_map.get(0), myOptions);
            Histrix.gmaps['".$this->Datos->idxml."'].myMarkers = [];

   
        	var transitLayer = new google.maps.TransitLayer();
        	transitLayer.setMap(Histrix.gmaps['".$this->Datos->idxml."']);
        /*	var panoramioLayer = new google.maps.panoramio.PanoramioLayer();
        	panoramioLayer.setMap(Histrix.gmaps['".$this->Datos->idxml."']);
            */	
    

            Histrix.gmaps['".$this->Datos->idxml."'].autoPan = true;
	
        	google.maps.event.addListenerOnce(Histrix.gmaps['".$this->Datos->idxml."'], 'dragend', function(){
        	    Histrix.gmaps['".$this->Datos->idxml."'].autoPan = false;
        	});

            Histrix.gmapSearched = false;
	        Histrix.searchPlaces = function(mylatlng){
	        var request = {
    	        location: mylatlng,
                radius: '20000',
                types: ['gas_station']
            };
                
            service = new google.maps.places.PlacesService(Histrix.gmaps['".$this->Datos->idxml."']);
            service.search(request, 
                function (results, status) {
                    if (status == google.maps.places.PlacesServiceStatus.OK) {
                        for (var i = 0; i < results.length; i++) {
                            var place = results[i];
                            createMarker(place, Histrix.gmaps['".$this->Datos->idxml."']);
                        }
                    }
                }
            );
                    

            function createMarker(place, map) {
                var placeLoc = place.geometry.location;
	       		var gpmarker = new google.maps.MarkerImage(place.icon, null, null, null, new google.maps.Size(25, 25));
                            
                var marker = new google.maps.Marker({
                    map: map,
	               	icon: gpmarker,
				    title: place.name,
				    position: place.geometry.location
                });
                                                              
                google.maps.event.addListener(marker, 'click', function() {
				    var infowindow = new google.maps.InfoWindow({
				        content: '<div id=\"place\">'+place.name+'</div>'
    				});
                    infowindow.open(map, marker);
                });
	
		    }
	    }

	    Histrix.gmaps['".$this->Datos->idxml."'].draw = function(){
            var data;

            // Return if removed
            if ($('#".$this->Datos->idxml."').length == 0){
                clearInterval(Histrix.gmaps['".$this->Datos->idxml."'].interval);
                return false;
            }

            jQuery.getJSON('process.php?reload=true&xmldatos=".$this->Datos->xml."&instance=". $this->Datos->getInstance()."', function(data){
                  var travel = [];
                  var myLatlng;
                  var titleData;
                  var count = 0;
                    /*
	        if (Histrix.gmaps['".$this->Datos->idxml."'].myMarkers.length){
                  
    		    $(Histrix.gmaps['".$this->Datos->idxml."'].myMarkers).each(function(){
    			this.setMap(null);
    		    });
                    
	    	    Histrix.gmaps['".$this->Datos->idxml."'].myMarkers.length = 0;
    	    }
                      */

            $.each(data, function(key, val) {
                count++;
			    
                var myPoint = val.point.split(',');
                myLatlng = new google.maps.LatLng(myPoint[0],myPoint[1]);
                titleData = val.title;

		      if (
    			count % 10 == 0 || 
    			//val.customData.gps_speed <= 1 || 
	   		    count == 1){
			         var infowindow = new google.maps.InfoWindow({
			             content: '<div id=\"content\">'+val.title+'</div>'
    			     });

                    if (Histrix.gmaps['".$this->Datos->idxml."'].myMarkers[count] != undefined)
			          Histrix.gmaps['".$this->Datos->idxml."'].myMarkers[count].setMap(null);
		    
                      var marker = new google.maps.Marker({
	                  position: myLatlng,
    	                  contentData:titleData,
    	                  title:'Click!',
/*    	                  animation: google.maps.Animation.DROP,*/
        	          map: Histrix.gmaps['".$this->Datos->idxml."'],
			  icon: {
			      path: google.maps.SymbolPath.FORWARD_CLOSED_ARROW,
			      rotation: parseInt(val.customData.gps_direction),
                    	      strokeColor: \"#aa1100\",

			      scale:2
		            }
            
            	      });
		    
		      if (count == 1 ){
//		        marker.animation =  google.maps.Animation.BOUNCE;
//		        marker.icon = \"../img/blue-dot.png\";
//		        marker.icon =\"thumb.php?url=../database/saldivia/xml//fotos/coches/1153/corte/Corte-1153-Alenco.pdf&ancho=50\";
			  marker.icon= {
			      path: google.maps.SymbolPath.FORWARD_CLOSED_ARROW,
			      rotation: parseInt(val.customData.gps_direction),
                    	      strokeColor: \"#0000AA\",

			      scale:3
		            };
		        if (!Histrix.gmapSearched) {
                            Histrix.searchPlaces(myLatlng);
                            Histrix.gmapSearched = true;
                        }
                            
			if (Histrix.gmaps['".$this->Datos->idxml."'].autoPan)
			        Histrix.gmaps['".$this->Datos->idxml."'].panTo(myLatlng);
		      } 

			Histrix.gmaps['".$this->Datos->idxml."'].myMarkers[count] = marker;

			google.maps.event.addListener(marker, 'click', function() {
			    var geocoder = new google.maps.Geocoder();
    	                    geocoder.geocode({'latLng': marker.position}, function(results, status){
	        		if (status == google.maps.GeocoderStatus.OK) {
		            	    if (results[0]) {                      
		            		infowindow.setContent( marker.contentData +'<hr/>' +results[0].formatted_address);
		            	    }
				}
    	                    });
			
		    	    infowindow.open(Histrix.gmaps['".$this->Datos->idxml."'] ,marker);
			});            	      
		    }

		    travel.push(myLatlng);
                 
                  });
		  var lineSymbol = {
		    path: google.maps.SymbolPath.BACKWARD_CLOSED_ARROW
		  };
                  
                  if (Histrix.gmaps['".$this->Datos->idxml."'].mytravelPath)
                      Histrix.gmaps['".$this->Datos->idxml."'].mytravelPath.setMap(null);
                  
                  Histrix.gmaps['".$this->Datos->idxml."'].mytravelPath = new google.maps.Polyline({
                      path: travel,
                      strokeColor: \"#FF0000\",
                      strokeOpacity: 1.0,
                      /*
                      icons: [{
                        icon: lineSymbol,
                        offset: '50%'
                      }],
                        */          
                      strokeWeight: 2
                     });
                                    
                  Histrix.gmaps['".$this->Datos->idxml."'].mytravelPath.setMap(Histrix.gmaps['".$this->Datos->idxml."']);
		    

                  
            });
            
        };
	
        // add listener after initial Load
        google.maps.event.addListenerOnce(Histrix.gmaps['".$this->Datos->idxml."'], 'idle', function(){

                      

                                                                                                        

            Histrix.gmaps['".$this->Datos->idxml."'].draw;
            clearInterval(Histrix.gmaps['".$this->Datos->idxml."'].interval);
	       Histrix.gmaps['".$this->Datos->idxml."'].interval = setInterval(  Histrix.gmaps['".$this->Datos->idxml."'].draw, 10000); // setTimeout
	
          });
       }; ";



        return $jscode;
    }


    // render de complete XML
    public function show($idFormulario = '', $divcont='', $opt='') {

        $id = 'Show' . $this->Datos->idxml;

        // id del contenedor (creo)
        $id2 = str_replace('.', '_', ($divcont != '') ? $divcont : $id);

        $style = (isset($this->Datos->style)) ? $this->Datos->style : '';

        // Columns
        if ($this->sidePanel === true)
            $ancho = (isset($this->Datos->ancho)) ? $this->Datos->ancho : '70';
        else
            $ancho = 100;
            
        $width = (isset($this->Datos->width)) ? $this->Datos->width : $ancho;

        if ($width != null) {
            $this->Datos->col1 = $width;
            $this->Datos->col2 = 100 - $width;
            $style.='width:' . $this->Datos->col1 . '%;';
        }
        $clasedetalle = 'detalle2';
        $clase = $this->defaultClass;
        $outputDatos = $this->showTabla();

//            if ($sidePanel)
 //       $barraSlide = $this->showSlider($id);


        $clase_der = (isset($this->formClass)) ? 'class="' . $this->formClass . '"' : 'class="consulta_der"';

    //    $outputAbm = $this->showAbm(null, $clase_der);


        // Si se define explicitamente una clase en el xml
        if ($this->Datos->clase != '') {
            $clase = $this->Datos->clase;
        }

        $retorno = '';
        if ($this->Datos->campoRetorno != '') {
            $uidRetorno = $this->Datos->getCampo($this->Datos->campoRetorno)->uid;
            $retorno = ' origen="' . $uidRetorno . '" ';
        }
 
        $output = '<div  class="' . $clase . '" id="' . $id . '" style="' . $style . '" ' . $retorno . '>';
        $output .= '<div class="contewin" >';
        $output .= $outputDatos;
        $output .= '</div>';
        $output .= '</div>';
        
        // El Abm
        $output .= $outputAbm;

        // Incorporo la barra vertical para slide
        $output .= $barraSlide;
        /*
        if ($this->sidePanel === true)
            $output .= $this->sidePanel();
         */
        
        // Add Detail div
   //     $output .= $this->detailDiv($clasedetalle);


        // create Javascript functions
        $script[] = "Histrix.registroEventos('" . $this->Datos->idxml . "')";
//        $script[] = "Histrix.calculoAlturas('" . $this->Datos->idxml . "', null ); ";
        $output .= Html::scriptTag($script);
        return $output;
    }

    public function showTablaInt($opt = '', $idTabla = '', $segundaVez = '', $nocant='', $div=false, $form=null, $pdf=null, &$parentObject=null) {
  /*      $llenoTemporal = (isset($this->Datos->llenoTemporal)) ? $this->Datos->llenoTemporal : '';
        $preload = (isset($this->Datos->preloadData)) ? $this->Datos->preloadData : '';

        if ($llenoTemporal != "false" && $segundaVez == '' && $opt != 'noselect') {
            if ($this->nosel != 'true') {
                if (!isset($this->Datos->tempTables[$llenoTemporal])) {

                    if ($preload != "false") {
                        $this->Datos->Select();
                    }
                    $this->Datos->preloadData = "true";

                    if ($this->Datos->resultSet)
                        $this->cantCampos = _num_fields($this->Datos->resultSet);
                }
                $this->Datos->CargoTablaTemporal();
            }
        }
*/
//        $this->createEvents();


        foreach($this->javascriptLibs as $n => $js) {
              $output .= Html::Tag('script', '',array('type' => 'text/javascript', 'src'=>$js ));  
        }

        $jscode[] = $this->googleMapsInit($this->Datos->idxml);

        $output .= Html::scriptTag($jscode);

        echo $output;
    }

    private function sidePanel() {
        $title = date('D, d M Y');
        $style = '';
        $strStyle = '';
        if ($this->Datos->col2 != '')
            $style .='width:' . ($this->Datos->col2 - 0.5) . '%;';

        if ($style != '')
            $strStyle = ' style="' . $style . '" ';


        $output .= "<div $strStyle class=\"consulta_der\">";
        //    $output .= "<h1>$title</h1>";
        $output .= "<div class=\"panel\">";
        $output .= "<ul id=\"eventList{$this->Datos->idxml}\" class=\"eventList\" ><li>Procesando..</li></ul>";
        $output .= '</div>';
        
        if ($this->Datos->calendarTasks != '') {
            $output .= "<div id=\"tasks{$this->Datos->idxml}\" class=\"panel\">";
            $output .= '</div>';
        }
         
        $output .= '</div>';
        return $output;
    }


    public function createEvents($startParam='',$endParam='') {

        $this->Datos->Select();
        $this->Datos->CargoTablaTemporal();

        $Tablatemp = $this->Datos->TablaTemporal->datos();
        $fieldlist = $this->Datos->camposaMostrar();

      //  $this->getDataFields($fieldlist);
        $separator = '';
        
        if (is_array($Tablatemp))
        {
            foreach ($Tablatemp as $rowNumber => $row) {
                $title = '';
                foreach ($row as $fieldName => $value) {
                    $field = $this->Datos->getCampo($fieldName);
                    if ($field->TipoDato == 'geoPoint'){
                        $point = $value;
                    }
                    else {
                	if ($field->noshow != 'true'){
                            $title .= $field->Etiqueta.': '.$value.'</br>';
                        }
                    }
                }

            $event['customData'] = $row;

            $event['title'] = $title;
            $event['point'] = $point;

            $this->eventArray[] = $event;


            }   

        }    

        $eventstr = json_encode($this->eventArray);

        return $eventstr;
    }


}

?>
