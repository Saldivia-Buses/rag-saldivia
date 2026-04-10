/**
 * HISTRIX javascript Library
 */

// JSLint paramenter
/*global $, jQuery, document, window, console, XulMenu, confirm*/

// from Firebugx.js
if (!window.console 
//|| !console.firebug
) {
	var names = [ "log", "debug", "info", "warn", "error", "assert", "dir",
			"dirxml", "group", "groupEnd", "time", "timeEnd", "cocunt",
			"trace", "profile", "profileEnd" ];
	window.console = {};
	var i = 0;
	for (i = 0; i < names.length; ++i) {
		window.console[names[i]] = function () {};
	}
}

var Histrix = {
	// histrix local variables

	readMsg : 120, // read message frequency
	readMail : 600, // read mail frequency
	currentElement : null,
	buscando : false,
	title_i : "0",
	message : "",
	temptitle : "",
	desktopNotifications : [],
	win : [],
	map : [],
	gmaps : [],
	polygonControl : [],
	markerControl : [],
	pluginHooks : {},
	menu1 : null,
	menu2 : null,
	db : '[]',
	onIdle:null,
	onIdleStop:null,	
	// Histrix local Objects.;
	fontSize : $('#fontSize'), // Font Size
	tabs : $('#tabs'), // Tabs
	//utilbar : $('#utilbar'), // Util Bar
	Supracontenido : $('#Supracontenido'), // Main div
	Msg : $('#Msg'), // Msg

	/**
	 * Init Histrix add menubar font Slide init messages
	 */
	init : function init(opener, options) {

		this.opener = opener;


		// current window ID
		this.uniqid=Math.random().toString(36).substr(2);;


		// datepicker options
		$.datepicker.setDefaults($.datepicker.regional[this.lang]);
		$.datepicker.setDefaults({
			showOn : 'button',
			buttonImageOnly : true,
			buttonImage : '../img/cal.gif',
			buttonText : 'Calendar',
			showAnim : 'drop',
			showOtherMonths : true,
			showButtonPanel : true,

			/* selectOtherMonths: true, */
			changeMonth : true,
			changeYear : true
		});

		// hack to gotoToday
	    /*jslint nomen: true*/

		$.datepicker._gotoToday = function (id) {
			var target = $(id),
			    inst = this._getInst(target[0]);
			if (this._get(inst, 'gotoCurrent') && inst.currentDay) {
				inst.selectedDay = inst.currentDay;
				inst.drawMonth = inst.selectedMonth = inst.currentMonth;
				inst.drawYear = inst.selectedYear = inst.currentYear;
			} else {
				var date = new Date();
				inst.selectedDay = date.getDate();
				inst.drawMonth = inst.selectedMonth = date.getMonth();
				inst.drawYear = inst.selectedYear = date.getFullYear();
				this._setDateDatepicker(target, date);
				this._selectDate(id, this._getDateDatepicker(target));
			}
			this._notifyChange(inst);
			this._adjustDate(target);
		};
	    /*jslint nomen: false*/

		var defaultOptions = { confirmExit : true },
		    histrixOptions = $.extend(defaultOptions, options);

		$('a[rel*=lightbox]').lightBox();

		$('.fileManager , .dropfile, .dropfile div').htxfiledrop();

		// Activate user List
		$('#msgBar').click(function () {
			$('#mensajeria').toggle();
		});

		this.initSlider();    // horizontal slider control
		//this.initSidePanel(); // SidePanel event initialitation
		this.initMessages();  // Messaging initialitation

		// automatic Menu Hiding hack
		 
		$('#Supracontenido, .htabbar, .hmenubar').not('.htrayclass, .hmenugradient').live('click',

		    function (event) {

				Histrix.hideMenu(event);
				$('.active').css({'display': 'none'}).toggleClass('active');
			}
		    );
		
		$('#barraestado').dblclick(function () {
			Histrix.sysmon();
		});

		if (this.opener) {
			var that = this;
			$('#alerta', that.opener.document).html(Histrix.i18n['sessionStart']);
			// .css('display', 'none');
		}
		window.onunload = function () {
			var that = this;
			if (that.opener) {
				$('#alerta', that.opener.document).html(Histrix.i18n['sessionClose']);
				// .css('display',
			}
		};

		$(window).resize(function () {
			Histrix.resizeAll();
		});

		if (histrixOptions.confirmExit) {
			window.onbeforeunload = function () {
				return false;
			};
		}

		// TimeOut
		// idleTimer() takes an optional argument that defines the idle timeout
		// timeout is in milliseconds; defaults to 30000
		$.idleTimer(90000);

		$(document).bind("idle.idleTimer", function () {
			// function you want to fire when the user goes idle
			Histrix.sysmon({
				'idle' : true
			});

			// Activate idle Plugin ex: marquee for anounces
			if (Histrix.onIdle != undefined && Histrix.onIdle != null ){
				Histrix.onIdle();
			}
			
		});

		$(document).bind("active.idleTimer", function () {
			// function you want to fire when the user becomes active again
			Histrix.sysmon({
				'idle' : false
			});
			if (Histrix.onIdleStop != undefined && Histrix.onIdleStop != null ){
				Histrix.onIdleStop();
			}
		});

		// Init Menu
		if ($("#menu1")[0]) {
			this.menu1 = new XulMenu("menu1");
			this.menu1.arrow1 = "../img/arrow1.gif";
			this.menu1.arrow2 = "../img/arrow2.gif";
			this.menu1.init();

		}
		
		if ($("#menu2")[0]) {
			this.menu2 = new XulMenu("menu2");
			this.menu2.type = "vertical";
			this.menu2.init();
		}
 
		$('.XulMenu a.item[rel], .XulMenu a.item[rel] span').live('click',
			function () {
				var menuItem = getParent(this, 'A'),
					$this = $(menuItem),
					url   = $this.attr('rel'),
				    myURL = parseURL(url),
				    xmlid = encodeURIComponent(url);

				if (myURL.params.xml != '') {
					xmlid = 'DIV' + myURL.params.xml;
				}
				var parsedxml = xmlid.replace(".", '_');
				xmlLoader(parsedxml, url, {
					title : $this.html(),
					menuId : $this.attr('menuId'),
					loader : $this.attr('loader'),
					helplink : $this.attr('helplink')
				  }
				);

				$('.item').removeAttr('last');
				$(menuItem).attr('last', 'true');

				Histrix.menu1.hideAll();
				 
				$('.menupanel').css({'display': 'none'});
				//Histrix.hideMenu(event);
				


			}
			);

		$('.menu_extra').each(
			function () {
				var $helpspan = $(this),
					helppath  = $helpspan.attr('hlppath');
				if (helppath) {
					$helpspan.addClass('boton').css({
						color : 'green'
					}).click(
						function (e) {
                                                        
							Histrix.loadInnerXML('prev4d95d05bf40f5',
									'pdfviewer.php?ro=true&f='
											+ helppath
											+ '&ancho=650&alto=500',
									null, $helpspan.text(), null, null,
									{
									width : '80%',
									height : '98%',
									modal : true
								} 

								);
								Histrix.menu1.hideAll();
								e.stopPropagation();
								
						}
					);
				}
			}
		);

		// Register live Editing
		// Automatic Save Changes for Live Grids
		$('.liveGridClass input , .liveGridClass select').live(
			'blur, change',
			function () {

				var table      = getParent(this, 'TABLE'),
					tr         = getParent(this, 'TR'),
					$tr        = $(tr),
					rowNumber  = $tr.attr('o'),
					insert     = $tr.attr('insert'),
					xml        = $(table).attr('xml'),
					instance   = $(table).attr('instance'),
					xmlcontext = $(table).attr('xmlOrig'),
					action     = 'update',
					vars       = Histrix.getRowValues(tr),
					oldvars    = Histrix.getRowValues(tr, null, 'esclave');
				/*
				 * if (insert == 'true') action = 'insert';
				 */
				$tr.attr('insert', 'false');

				Histrix.showMsg(action);
				var url = "process.php?getFila=true&fila=" + rowNumber
						+ "&xmldatos=" + xml + '&instance=' + instance
						+ '&accion=' + action;
				var postData = {
					newValues : vars,
					oldValues : oldvars,
					xmlOrig : xmlcontext,
					'__winid':Histrix.uniqid
				};

				$.post(url, postData, function () {
					Histrix.hideMsg();
				});
				// Histrix.hideMsg();

			}
		);

		$('.liveGridClass tr .deleteImageButton').live(
			'click',
			function () {
				var message = 'Borrar el Registro?';
				if (confirm(message)) {
					var table = getParent(this, 'TABLE'),
						tr = getParent(this, 'TR'),
						$tr = $(tr),
						rowNumber = $tr.attr('o'),
						xml = $(table).attr('xml'),
						instance = $(table).attr('instance'),
						xmlcontext = $(table).attr('xmlOrig'),
						action = 'delete',
						vars = Histrix.getRowValues(tr),
						oldvars = Histrix.getRowValues(tr, null, 'esclave');

					Histrix.showMsg(action);
					$.post("process.php?Nro_Fila=" + rowNumber
							+ "&instance=" + instance + "&xmldatos=" + xml
							+ '&accion=' + action, {
						newValues : vars,
						oldValues : oldvars,
						xmlOrig : xmlcontext,
						'__winid':Histrix.uniqid
					}, function () {
						Histrix.hideMsg();
						$tr.remove();

					});
				}
			});

		// Internal Link 
		$('.internalLink').live('click',function (e){
			var attrHref = $(this).attr('href');

			var href = 'cargahttp.php?url='+ attrHref.replace("?", '&');
			Histrix.loadInnerXML('internalLink', href, options, $(this).text() , undefined, undefined, {modal:true});
			//, xmlprincipal, idobj, winoptions);
			return false;

		});
		// messages
		$('.send_msg, #_usroffline li').live(
				'click',
				function (e) {

					var $target = $(e.target).closest('[login]');

					var login = $target.attr('login');
					if (login) {
						var Nom = encodeURIComponent($target.html());
						var email = encodeURIComponent($target.attr('email'));
						var telefono = encodeURIComponent($target
								.attr('telefono'));

						// request notification permissions
						if (window.webkitNotifications)
							window.webkitNotifications.requestPermission();

						Histrix.loadInnerXML('mensajes_reply_xml',
								'histrixLoader.php?xml=mensajes_reply.xml&destino='
										+ Nom + '&Email=' + email + '&para='
										+ login + '&telefono=' + telefono
										+ '&dir=histrix/mensajeria', '',
								'Enviar', 'DIVmensajes_recive.xml', 'Enviar', {
									width : '500px',
									height : '320px'
								});
					}
				});

		// raise hidden windows
		$('div.ventint.ui-draggable').live('click', function (e) {
			if (e.target.tagName != 'BUTTON') {
				$window = $(this).closest('.ui-draggable');
				$('.ventint.ui-draggable').css('z-index', '1');
				$window.css('z-index', '8');
			}
		})

		$('.dragBar').live('click', function (e) {
			var idxml = this.id.substring(7);
			var button = $(e.target).closest('button');

			if (button.hasClass('maximizeButton')) {
				Histrix.maximize(idxml);
				e.stopPropagation();
			}
			if (button.hasClass('closeButton')) {

				cerrarVent(idxml, button);
				e.stopPropagation();
			}
		});

		// Fill From Event
		$('table.sortable[fillform=true], table.Tree').live(
				'click',
				function (e) {
					var $target = $(e.target); // e.target grabs the node that
												// triggered the event.
					var tr = $target.parent('tr');

					var drag = $target.closest('td.dragHandle');
					var detailCell = $target.closest('td.detailCell');

					if (drag.length > 0 || detailCell.length > 0)
						return false;

					// if (tr[0] && tr[0].hasAttribute('o') &&
					// !tr[0].hasAttribute('block')){
					if (tr[0] && (tr[0].getAttribute('o') != undefined)
							&& tr[0].getAttribute('block') == undefined) {

						e.stopPropagation(); // prevents further bubbling
						fillForm(tr[0], null);

						// custom row event
						// after fillform
						var customTRjs = tr.attr('onRowClick');
						if (customTRjs) {
							eval(customTRjs);
						}

					}
				});

		// Load Detail detail

		$('table.sortable tr[detailpar], table.Tree tr[detailpar], ul.ullist li[detailpar] ')
				.live(
						'click',
						function (e) {

							var $targetElement = $(e.target);
							var tr = $targetElement.closest('[o]')[0];
							var cell = tr;
							var options = {};

/*							if ($targetElement.is('a') || $targetElement.is('button') || 
							    $targetElement.parent().is('button')) {
							    //return false;
							    // prevents double trigger
							} else 
*/
							if ($targetElement[0].getAttribute('detailcell') != undefined
									|| $targetElement.hasClass('detailCell')) {
									
									
								    
								cell = $targetElement.closest('td')[0];
                                                                                                  /*
								    if ($(cell).hasClass('single')){
									$table = $targetElement.closest('table');
									                          
									$('tr[openrow=true]', $table).each(
									    function (){
									    
										var $tr2 = $(this);

									    if ($(cell) != $tr2)
    								Histrix.cargoDetalle('histrixLoader.php?'
										+ $tr2.attr('detailpar'), $tr2[0],
										$tr2.attr('detaildiv'), {inline:true});

									    }
									);
								    }
								                                    */
								
								
								options = {
									inline : true
								};
							}

							if (tr.getAttribute('detailpar') != undefined) {
								// if (tr.hasAttribute('detailpar')){
								// e.stopPropagation(); // prevents further
								// bubbling
								Histrix.cargoDetalle('histrixLoader.php?'
										+ tr.getAttribute('detailpar'), cell,
										tr.getAttribute('detaildiv'), options);
							}
						});
		// double click event for tablets
		$('table.sortable tr[detailpar], table.Tree tr[detailpar]').live(
				'dblclick',
				function (e) {

					var $targetElement = $(e.target);
					var tr = $targetElement.closest('tr')[0];
					var cell = tr;
					var options = {};
					if ($('.detailCell', tr)[0] != undefined) {
						cell = $('.detailCell', tr)[0];
						options = {
							inline : true
						};
					}
					if (tr.getAttribute('detailpar') != undefined) {
						Histrix.cargoDetalle('histrixLoader.php?'
								+ tr.getAttribute('detailpar'), cell, tr
								.getAttribute('detaildiv'), options);
					}
				});



		// button Events
		$('table.sortable, table.Tree, table.form, div.card, li.list, div.filtro, .botoneraImp')
				.live(
						"click",
						function (e) {

							var triggerObj = $(e.target);

							var table = triggerObj.closest('table')[0];
							var addRowButton = triggerObj.closest('button');

							// SELECT CURRENT ROW
							$('.trsel', table).removeClass('trsel');
							var tr = triggerObj.parent('tr');
							$(tr).addClass('trsel');

							// Add Row to table
							if (addRowButton.hasClass('addRow')) {
								Histrix.addRow(table);

								e.stopPropagation();
							}
							if (triggerObj.is('a') || triggerObj.is('button') || triggerObj.parent().is('button')) {

								// THIS FIX IMAGES INSIDE!!!!!!!!!!!:
								if (triggerObj.parent().is('button')){
								    triggerObj = triggerObj.parent();
								}

								if (triggerObj[0].getAttribute('linktarget') != undefined) {
									// if
									// (triggerObj[0].hasAttribute('linktarget')){

									var target = triggerObj.attr('linktarget');
									var menuId = '&_menuId='
											+ triggerObj.attr('menuid');
									var loader = 'histrixLoader.php?'
											+ triggerObj.attr('linkloader')
											+ menuId;
									e.stopPropagation(); // prevents further
															// bubbling
									switch (target) {
									case 'win':
										var position = $(triggerObj).position();
										var winoptions = {};
										if (triggerObj.attr('linkreposition')) {
											winoptions['posY'] = position.top;
											winoptions['posX'] = position.left
													+ $(triggerObj)
															.outerWidth();
										}

										if (triggerObj.attr('linkwidth')) {
											winoptions['width'] = triggerObj
													.attr('linkwidth');
										}
										if (triggerObj.attr('linkheight')) {
											winoptions['height'] = triggerObj
													.attr('linkheight');
										}
										if (triggerObj.attr('linkmodal')) {
											winoptions['modal'] = triggerObj
													.attr('linkmodal');
										}
										Histrix.loadInnerXML(triggerObj.attr('linkint'), loader, '',
												triggerObj.attr('title'),
												triggerObj.attr('linkfather'),
												triggerObj.attr('linkfname'),
												winoptions);
									e.stopPropagation(); // prevents further

										return false;


										break;
									case 'tab':
										Histrix.loadXML(triggerObj
												.attr('linkint'), loader, '',
												triggerObj.attr('title'),
												triggerObj.attr('linkfather'));
										break;
									case 'print':
										// Direct Printing, choose printer
										Histrix.printExt(loader, 'true',
												triggerObj);

										break;
									}
								}
							}

						});

		 // filter live form events //
		$('form.autofilter :input').live(
		    'change',
		    function (e){
			$('.searchButton' ,$(this).closest('form')).click();
		    }
		);
		
		
		$('body').live(
				'keydown',
				function (event) {
				
					var triggerObj = $(event.target);
					var keyCode = event.keyCode ? event.keyCode
							: event.which ? event.which : event.charCode;

					switch (keyCode) {
					case 120: // F9
						if($('.ui-widget-overlay').length == 0) // prevents double fire
						$('[name=Procesar]:visible:last').each(function () {

							var $button = $(this);
							if ($button.is(':enabled') == true) {
								$button.click();
							}
						});
						break;
					case 27: // ESC
						$('.closeButton:visible:last, .barraBusqueda button').click();

						break;
					case 113: // HELP
											
						if (triggerObj.is(':input[helpData]')) {
						
							var helpData = jQuery.parseJSON( triggerObj.attr('helpData'));

							popAyuda( 'DIV'+ helpData.xmlform , triggerObj  , helpData.xmlHlp,  event,  helpData );							

							event.stopPropagation(); // prevents further bubbling
							return false;
						}
						
						if (triggerObj.closest('form').is('[searchForm]')){
							
							ayudaFicha( triggerObj);	
						}

						break;
					case 115:
						if (triggerObj.closest('form').is('[searchForm]')){
							var $form = triggerObj.closest('form');
							Histrix.clearForm($form.attr('xml'), true);
						}						
					break;
					}

					if (triggerObj.is(':input')) {
						if (triggerObj[0].tagName != 'TEXTAREA'
								&& triggerObj[0].tagName != 'BUTTON') {
						 triggerObj.keypress(Histrix.handleEnter);

						}
					}

				});
		 
		// Remove wait
		$('#waitScreen').remove();

	},

	osTray: function osTray(options){

		if (Histrix.Unity){
			Histrix.Unity.Launcher.setCount(options.count);
			if (options.count == 0){
				Histrix.Unity.Launcher.clearCount();
			}
		}
	},
	/**
	 * Register PluginHooks into Histrix.pluginHooks Object
	 * 
	 * @param HookName -
	 *            String
	 * @param functionName -
	 *            Function to be executed
	 */
	registerHook : function registerHook(hookName, fnName) {
		var currentHooks = this.pluginHooks[hookName];
		if (currentHooks) {
			currentHooks.Push(fnName);
			this.pluginHooks[hookName] = currentHooks;
		} else {
			var hooks = new Array(fnName);
			this.pluginHooks[hookName] = hooks;
		}
	},

	destroy : function destroy(motive) {
		$('#utilbar, .htabbar, #menubar, #barraestado, #ultimosprogs').remove();
		this.Supracontenido
				.html('<div class="error" style="margin: 20%; padding: 10px; text-align: center; font-size: 12px; font-weight: 700;"><b>'
						+ motive + '</b></div>');
	},

	/**
	 * Side Panel Events
	 */
	/*
	initSidePanel : function initSidePanel() {
		// Utility Bar Events
		var Histrix = this;
		$('.utilbarStatus').click(Histrix.toggleUtilBar);
		this.utilbar.dblclick(Histrix.toggleUtilBar);
	},
	 */
	/**
	 * Font size slider
	 */
	initSlider : function initSlider() {
		$('#fontSlider').slider({
			value : 12,
			min : 7,
			max : 40,
			step : 1,
			slide : function (event, ui) {
				$("#fontSize").html(ui.value);
				Histrix.changeFont(ui.value);
			}
		});
	},

	resizePanel : function resizePanel (uidSlide) {
		var $slide = $('#' + uidSlide);

		var divDerecho = $slide.attr('rdiv');
		var divDerecho2 = $slide.attr('rdiv2');

		var id = $slide.attr('idxml');
		$slide.draggable({
			axis : 'x',
			drag : function (event, ui) {
				var x = ui.position.left;
				$('#' + id).css("width", x + "px");

				$('#' + divDerecho).css({
					"left" : x + 8 + "px",
					"right" : 0 + "px",
					"width" : "auto"
				});
//				if (divDerecho2 != '') {
					$('#' + divDerecho2).css({
						"position" : "absolute",
						"left" : x + 8 + "px",
						"right" : 0 + "px",
						"width" : "auto"
					});
//				}
			}
		});

		$slide.dblclick(function (ind, ui) {
			$('#' + uidSlide).css("left", "50%");
			var pos = $('#' + uidSlide).position();
			var x = pos.left;

			$('#' + id).css("width", x + "px");
			$('#' + divDerecho).css({
				"left" : x + 8 + "px",
				"right" : 0 + "px",
				"width" : "auto"
			});

			if (divDerecho2 != '') {
				$('#' + divDerecho2).css({
					"position" : "absolute",
					"left" : x + 8 + "px",
					"right" : 0 + "px",
					"width" : "auto"
				});
			}
		});

		var x = $slide.position().left;

		if (x != 0) {
			$('#' + id).css("width", x + "px");
			$('#' + divDerecho).css({
				"left" : x + 8 + "px",
				"right" : 0 + "px",
				"width" : "auto"
			});

			if (divDerecho2 != '') {
				$('#' + divDerecho2).css({
					"position" : "absolute",
					"left" : x + 8 + "px",
					"right" : 0 + "px",
					"width" : "auto"
				});
			}
		}
	},

	updateWebmail : function updateWebmail() {
		$.ajax({
			url : "mail.php",
			async : true,
			timeout : 7000,
			success : function (html) {
				$("#webmail").html(html);
				setTimeout(Histrix.updateWebmail, Histrix.readMail * 1000);
			}
		});

	},
	periodicalUpdate : function periodicalUpdate(cont, data) {
		$(cont).load(
				data.url,
				{'__winid':Histrix.uniqid},
				function (responseText, textStatus, XMLHttpRequest) {
					setTimeout("Histrix.periodicalUpdate('" + cont
							+ "', {url:'" + data.url + "' , minTimeout:"
							+ data.minTimeout + "});", data.minTimeout * 1000);
				});
	},
	sysmon : function sysmon(postData) {
		//spostData['__winid'] =Histrix.uniqid;
		$('#mensajeria').load('sysmon.php', postData);
	},
	updateMensajeria : function updateMensajeria() {
		$('#mensajeria')
				.load(
						'sysmon.php',
						{'__winid':Histrix.uniqid},
						function (responseText, textStatus, XMLHttpRequest) {
							setTimeout(Histrix.updateMensajeria,
									Histrix.readMsg * 1000);
						});
	},
	/**
	 * Init internal messages check
	 */
	initMessages : function initMessages() {
		jQuery.ajaxSetup({
			async : true
		});
		Histrix.updateMensajeria();
		Histrix.updateWebmail();

		$('#mensajeria').bind('mouseover', function (e) {
			var login = $(e.target).attr('login');
			if (login) {
				Histrix.userClose();
				Histrix.userInfo(e.target, login);
			}
		});
		$('.ultimosconect, #userInfo').bind('mouseout', function (e) {
			Histrix.userClose();
		});

	},

	userInfo : function userInfo(elem, login) {
		$elem = $(elem);
		var Liemail = $elem.attr('email');
		var Lifoto = $elem.attr('foto');
		var Liprofile = $elem.attr('profile');
		var Liinterno = $elem.attr('interno');
		var Litelefono = $elem.attr('telefono');
		var Liname = $elem.text();
		var position = $elem.offset();
		var w = $elem.width() + 185;

		var pos = $elem.offset();
		var tempX = pos.left;
		var tempY = pos.top;

		var finalTop = tempY - elem.clientHeight;

		var userFoto = '';

		if (Lifoto != '')
			userFoto = '<img src="' + Lifoto + '" >';
		var h = document.body.offsetHeight;

		if (position.top + 100 > h)
			position.top = h - 100;

		var htmlData = '<div class="userOptions">Información</div><div class="userData">'
				+ '<div><b>'
				+ Liname
				+ '</b></div>'
				+ Liemail
				+ '<div>Interno: '
				+ Liinterno
				+ '</div>'
				+ '<div>teléfono: '
				+ Litelefono
				+ '</div>'
				+ '<div>Sección: '
				+ Liprofile
				+ '</div>'
				+ '</div>'
				+ '<div class="userPhoto">'
				+ userFoto
				+ '</div>';

		$('#userInfo').html(htmlData).addClass('shadow').css({
			// top: position.top - 2 + 'px',
			// left: position.left - w + 'px',
			top : finalTop + 'px',
			left : (tempX) - w + 'px',
			display : 'block',
			position : 'absolute',
			zIndex : 999
		}).bind('mouseover', function () {

			$(this).css({
				display : 'block'
			});
		});
	},
	userClose : function userClose() {
		$('#userInfo').css('display', 'none');
	},

	hideMenu : function hideMenu(e) {
		
		var targ;
		if (!e)
			e = window.event;
		if (e.target)
			targ = e.target;
		else if (e.srcElement)
			targ = e.srcElement;
		if (targ.nodeType == 3) // defeat Safari bug
			targ = targ.parentNode;
		if (targ.className != "item" && targ.className != "item-active"
				&& targ.className != "button-active") {
			if (this.menu1)
				this.menu1.hideAll();
		}

	},
	/**
	 * Register Form Events EVENTS xml: xml name of the Data Container
	 */
	registroEventos : function registroEventos(xml) {

		// reset form
		// WARNING it fixes inner forms create bug
		
		//Histrix.clearForm(xml, false);

		// Select all links that contains lightbox in the attribute rel
		$('a[rel*=lightbox]').lightBox();
		    
		// activate recurrence rules inputs
		$('.recurrenceinput_form').remove();


		$(".recurringrule", "#Form"+ xml).each(function (){
		    var rrfield = this;
		    // dynamic load script
		    if( !jQuery.isFunction( $.recurrenceinput )){

			$.requireScript(['../javascript/recurrenceinput/jquery.tmpl-beta1.js', 
					 '../javascript/recurrenceinput/jquery.tools.overlay-1.2.7.js', 
					 '../javascript/recurrenceinput/jquery.tools.dateinput-1.2.7.js',
					 '../javascript/recurrenceinput/jquery.utils.i18n-0.8.5.js',
					 '../javascript/recurrenceinput/jquery.recurrenceinput.js', 
					 '../javascript/recurrenceinput/jquery.recurrenceinput-'+ Histrix.lang +'.js'] , 
			    function (){
			    /*
$.tools.recurrenceinput.setTemplates(
            {
                daily: {
                    rrule: 'FREQ=DAILY',
                    fields: [
                        'ridailyinterval',
                        'rirangeoptions'
                    ]
                },
                mowefr: {
                    rrule: 'FREQ=WEEKLY;BYDAY=MO,WE,FR',
                    fields: [
                        'rirangeoptions'
                    ]
                },
                weekends: {
                    rrule: 'FREQ=WEEKLY;BYDAY=SA,SU',
                    fields: [
                        'rirangeoptions'
                    ]
                }
            }, 
            {
                en: {
                    daily: 'Daily',
                    mowefr: 'Mondays, Wednesdays, Fridays',
                    weekends: 'Weekends'
                },
                es: {
                    daily: 'Diario',
                    mowefr: 'Lunes, Miercoles, Viernes',
                    weekends: 'Fines de Semana'
                }
            }        
        ); 
        */
				   $(rrfield).recurrenceinput({lang:Histrix.lang});

			}, false);
		    }
		    else {
		    
                              //    alert('requirescript loaded');
                              /*
$.tools.recurrenceinput.setTemplates(
            {
                daily: {
                    rrule: 'FREQ=DAILY',
                    fields: [
                        'ridailyinterval',
                        'rirangeoptions'
                    ]
                },
                mowefr: {
                    rrule: 'FREQ=WEEKLY;BYDAY=MO,WE,FR',
                    fields: [
                        'rirangeoptions'
                    ]
                },
                weekends: {
                    rrule: 'FREQ=WEEKLY;BYDAY=SA,SU',
                    fields: [
                        'rirangeoptions'
                    ]
                }
            }, 
            {
                en: {
                    daily: 'Daily',
                    mowefr: 'Mondays, Wednesdays, Fridays',
                    weekends: 'Weekends'
                },
                es: {
                    daily: 'Diario',
                    mowefr: 'Lunes, Miercoles, Viernes',
                    weekends: 'Fines de Semana'
                }
            }        
        ); 
			        */
			$(rrfield).recurrenceinput({lang:Histrix.lang});		    
		    }
			
		});

		$('ul.sortablelist').sortable({
			placeholder : "ui-state-highlight",
			forcePlaceholderSize : true,
			cursor : 'crosshair',
			update : function (event, ui) {
				var xmldatos = $(this).attr('xml');
				var instance = $(this).attr('instance');
				var rowEnd = $('li', this).index(ui.item);
				// var rowEnd = row.rowIndex - 1;
				// loger(rowStart);
				// loger(rowEnd);
				// if (rowStart != row.rowEnd)
				$('#Msg').load('swapFilas.php', {
					'instance' : instance,
					'xmldatos' : xmldatos,
					'dndSource' : rowStart,
					'dndTarget' : rowEnd,
					'__winid':Histrix.uniqid
				});
				// loger($(ui.item));
				// $(ui.item).addClass('error');
				/*
				 * $(ui.item).pulse({ backgroundColors: ['yellow','orange'],
				 * runLength: 1, speed: 500 });
				 */
			},
			start : function (event, ui) {

				rowStart = $('li', this).index(ui.item);

				/*
				 * var row = $(td).parent('tr')[0]; rowStart = row.rowIndex - 1;
				 */
			}
		});

		// draggable inner forms
		$('.singleForm').draggable({
			handle : '.singleForm legend'
		});

		// apply change event to set default values
		$(".refreshable", "#Form" + xml).change();

		$('#' + 'Form' + xml + ' :input ,' + '#' + 'FForm' + xml + ' :input')
				.each(
						function () {
							this.xml = xml;
							var $obj = $(this);
							$obj.focus(Histrix.tooltip).blur(
									Histrix.removeCurrentElement).change(
									Histrix.calculo);
							/*
							 * if ($obj[0].tagName !='TEXTAREA' &&
							 * $obj[0].tagName !='BUTTON'){
							 * $obj.keypress(Histrix.handleEnter); }
							 */
							if ($obj.attr('escolor') != undefined) {
							    // dynamic load script
							    if(typeof attachColorPicker == "undefined"){
								$.requireScript('../javascript/colorpicker.js', function (){
									attachColorPicker($obj[0]);
								});
							    }
							    else{
								attachColorPicker($obj[0]);
							    }
							    
							}
						});

		$('[internal_class="simpleditor"]')
			.filter('[class!=editorActivated]')
			.addClass('editorActivated').each(function (){
			var $this = $(this);
			// dynamic load wysiwyg
			if (typeof wysiwyg == "undefined"){
			    $.requireScript('../javascript/jwysiwyg/jquery.wysiwyg.js', function (){
				$this.wysiwyg();
		    $this.wysiwyg("setContent");

			    });
			}
			else {
			    $this.wysiwyg();
		    $this.wysiwyg("setContent");

			}


		});
		
		$('textarea[maxlength]').maxLength();
		
		$('span[forceTittle]').each(
				function () {
					var $this = $(this);
	    			// dynamic load tinyTips
				    if (typeof tinyTips == "undefined"){
	        			    $.requireScript('../javascript/jquery.tinyTips.js', function (){

					$this.css('color', 'red').tinyTips('light',
							$this.parent().attr('valor'));
					});
				    }
				    else {
					$this.css('color', 'red').tinyTips('light',
							$this.parent().attr('valor'));
				    }
				});

		$(".SpinButton").each(function (index, elem) {
			var $this = $(elem);
			var Emin = $this.attr('min');
			var Emax = $this.attr('max');

			$this.SpinButton({
				min : Emin,
				max : Emax
			});
		});

		$(':input[mask]').each(function (e) {
			$this = $(this);
			$this.unmask();
			$this.mask($this.attr('mask'));

		});

		$(':input[defaultvalue]').click(function (e) {
			ghost(this, e);
		}).focus(function (e) {
			ghost(this, e);
		}).blur(function (e) {
			ghost(this, e);
		});


		$("input.date:enabled").datepicker();
		$('input.month:enabled')
				.datepicker(
						{
							changeMonth : true,
							changeYear : true,
							buttonImageOnly : false,
							showButtonPanel : true,
							dateFormat : 'yymm',
							onClose : function (dateText, inst) {
								var month = $(
										"#ui-datepicker-div .ui-datepicker-month :selected")
										.val();
								var year = $(
										"#ui-datepicker-div .ui-datepicker-year :selected")
										.val();
								$(this).datepicker('setDate',
										new Date(year, month, 1));
							},
							beforeShow : function (input, inst) {
								inst.dpDiv.addClass('month-calendar');

								if ((datestr = $(this).val()).length > 0) {
									year = datestr.substring(
											datestr.length - 4, datestr.length);
									month = jQuery
											.inArray(datestr.substring(0,
													datestr.length - 5),
													$(this).datepicker(
															'option',
															'monthNames'));
									$(this).datepicker('option', 'defaultDate',
											new Date(year, month, 1));
									$(this).datepicker('setDate',
											new Date(year, month, 1));
								}
							}
						});

		Histrix.resizeAll();

		var $tableFoot = $('#tfoot_'+xml);
		Histrix.tableTotals($tableFoot);
	},

	
	tableTotals: function tableTotals(tableFoot, field){
		$('[jsevaldest]', tableFoot).each(function(){

			var $this = $(this);

			var evaldest = jQuery.parseJSON($this.attr('jsevaldest'));
			var parentInstance = $this.attr('jsparent');

			var importe = $this.html();

			if (evaldest != undefined)
				for ( var k = 0; k < evaldest.length; k++) {

					var destino2 = $('[name="' + evaldest[k] + '"]', '[instance="'+ parentInstance+'"]');

					if (destino2[0] == undefined) {
						destino2 = $('[name="' + evaldest[k] + '"]');
					}

					if (destino2[0] == undefined)
						continue;

					var dest2 = destino2[0];

					if (dest2.tagName == 'INPUT') {

						dest2.value = importe;

						var xmldatos = $(dest2.form).attr('xml');
						// seteo el valor del campo en el contenedor
						// comentado NO ANDA la suma de las grillas.

						setCampo(evaldest[k], dest2.value, xmldatos, 'false', dest2);

					}

					if (dest2.getAttribute('jsevaldest') != undefined) {

						//var dests = 'var array_destinos= ' + dest2.getAttribute('jsevaldest');
						//eval (dests);

						var array_destinos = jQuery.parseJSON(dest2.getAttribute('jsevaldest'));
						var calc  = 'var array_calcular= ' + dest2.getAttribute('jseval');

						eval(calc);



						for ( var j = 0; j < array_destinos.length; j++) {
							var jseval = array_calcular[j];
							var jsevaldest = array_destinos[j];
							calculojs(jseval, jsevaldest, dest2.form.id, dest2,null, false, 'false');
						}
					}
					
					if (field) {
						if (act == true) {
							var xmlParent = $(field.form).attr('xml');
							if (tabla != undefined) {
								var miJSON = Histrix.updateGrid(tabla);
								var grilla = tabla.parentNode.parentNode.id; // rehacer
								var instance = $(tabla).closest('table').attr(
										'instance');
							}
							if (instance != undefined && grilla != undefined) {
								$.post("setData.php?_show=false"
										+ '&instance=' + instance + '&xmlOrig='
										+ xmlParent + "&actualizoTabla=true", {
									mijson : miJSON,
									'__winid':Histrix.uniqid
								});
							}
						}
					}
					
				}


			}
		);
	},

	/**
	 * ENTER KEY EVENT
	 */
	handleEnter : function (event) {
		var keyCode = event.keyCode ? event.keyCode : event.which ? event.which
				: event.charCode;
		// Makes ENTER key to perform as TAB (except for TextAreas
		if (keyCode == 13) {
			$(event.target).blur().focusNextInputField();
			return false;
		}
		return true;
	},

	titler : function (msg) {
		// loger(msg);

		if (msg == 'histrix') {
			document.title = '[' + Histrix.db + ']';
			return;
		}

		if (Histrix.message == msg) {
			return;
		}

		if (Histrix.message == msg) {
			return;
		}
		if (msg != undefined){
			Histrix.message = msg;

		
		}

		var speed = "150";
		document.title = Histrix.temptitle
				+ Histrix.message.charAt(Histrix.title_i);
		Histrix.temptitle = Histrix.temptitle
				+ Histrix.message.charAt(Histrix.title_i);
		Histrix.title_i++;
		if (Histrix.title_i == Histrix.message.length) {
			Histrix.title_i = "0";
			Histrix.temptitle = "";
		}

		if (Histrix.message != '')
			setTimeout(Histrix.titler, speed);
	},

	// //////////////////////////////////////
	// Field Tooltip
	// //////////////////////////////////////
	tooltip : function () {
		$("#tooltip" + this.xml).html(this.title);
		Histrix.currentElement = this.id;
	},

	removeCurrentElement : function () {
		// Update current Element
		Histrix.currentElement = null;
	},

	/**
	 * Show/Hide Objects
	 * 
	 * @param Element
	 *            Html Element
	 */
	toggle : function (Element) {
		// this.playSound('select');
		$(Element).toggle("normal");
	},

	/**
	 * Loading message
	 * 
	 * @param Element
	 *            Html Element
	 * @param message
	 *            Text Message
	 */
	loadingMsg : function (Element, message , loading) {
		if(loading != false){
		$('#' + Element).html(
				'<div class="esperareloj"><b>' + message
						+ '</b><div id="throbber" /></div>');
		} else {
		$('#' + Element).html(
				'<div class="esperareloj"><b>' + message
						+ '</b></div>');

		}
	},

	/**
	 * Register Table Events
	 * 
	 * @param xml
	 *            xml name of the Data Container
	 * @param className :
	 *            class name to bind events to
	 */
	registerTableEvents : function (xml, className) {

		var table = $("#" + xml);
		if (className)
			table = $("." + className);
		if (table) {
			var $table = $(table);
			// Drag And Drop
			var rowStart;
			if ($table.hasClass('dnd')) {

				$('tr', table).hover(function () {
					$(this.cells[0]).addClass('showDragHandle');
				}, function () {
					$(this.cells[0]).removeClass('showDragHandle');
				});
				var xmldatos = $table.attr('xml');
				var instance = $table.attr('instance');
				// var idxmldatos = xmldatos.replace(".", '_');
				$table
						.tableDnD({
							onDragClass : "dndRow",
							dragHandle : "dragHandle",
							onDrop : function (table, row) {

								var rowEnd = row.rowIndex - 1;

								if (rowStart != row.rowEnd) {

									Histrix.showMsg('Ordenando Tabla');
									$
											.post(
													'swapFilas.php?dndSource='
															+ rowStart
															+ '&dndTarget='
															+ rowEnd
															+ '&xmldatos='
															+ xmldatos
															+ '&instance='
															+ instance,
													{'__winid':Histrix.uniqid},
													// after post
													function () {
														$table
																.parent()
																.load(
																		"refrescaCampo.php?xmldatos="
																				+ xmldatos
																				+ "&nocant=true&select=false&instance="
																				+ instance,
																		{'__winid':Histrix.uniqid},
																		function () {
																			var $newtable = $("#"
																					+ xml);
																			if (className)
																				$newtable = $("."
																						+ className);
																			// reposition
																			// Row
																			positionRow(
																					$newtable,
																					rowEnd + 1,
																					'update');
																			Histrix.hideMsg();
																		});
													});

								}

								return false;

							},
							onDragStart : function (table, td) {
								var row = $(td).parent('tr')[0];
								rowStart = row.rowIndex - 1;
							}
						});
			}

			// live sum?
			// Initialize Table
			// Select First Row
			// rows = table.eq(0).rows;
			// var myTableRow= table.eq(0).rows[1];
			// $(myTableRow).addClass('trsel');
			$table.children("tr:first").addClass('trsel');
			if ($table[0])
				$(document).click(function (event) {
					Histrix.currentElement = $table[0].id;
				});
		}

	},

	copyTableFooter: function (table){
		var $table = $(table);
		var xml =  $table.attr('xml');
 
		if (xml){
		var idxml  = xml.replace(".", '_');

			var $originalFooter = $( 'tfoot tr', $table);

			var detailTD = $originalFooter.closest('td.detailTD');

			if (detailTD.length != 0) {
				return;
			}

			var footer = $originalFooter.clone();
			footer.css('color','#000');

			// duplicate widths
			$originalFooter.children().each(
				function (){
					var cellwidth = $(this).width();

					footer.children().eq(this.cellIndex).css('width', cellwidth+'px').removeClass('sintotal');
				});


			var $totales = $('#totales_'+ idxml);    	
			if (footer && $totales.length != 0){

    		//	$('#footer'+idxml).remove();

	    		footer.attr('id', 'footer'+idxml);

		    	// change ids
			    $('th',footer).each(function (){
				    this.id = 'cloned_'+this.id;
    			}).addClass('ui-widget-header cellsum');

                if ($('#footer'+idxml).length == 0)
					$totales.append(footer);
				// hide original Footer
				$originalFooter.css('display','none');


			}
		}
	},

	// trying to improve calcheight
	calculoAlturas : function (xml) {
		// alert('reflow '+xml);

		// Get Main Container
		var idxml = xml.replace(".", '_');
		var $main = $('#DIV' + idxml);
		var cabeceraIng = 0;
		var $detailmain = null;
		// Detail
		if ($main.length == 0) {
			$detailmain = $('#' + idxml);
			if ($detailmain.length > 0) {
				var target = $detailmain.attr('target');
				if (target)
					idxml = target.replace(".", '_');
			}
		}

		if ($main.length == 0)
			$main = $('#Show' + idxml);
		if ($main.length == 0)
			$main = $('#IMP' + idxml);

		if ($('#INTER' + idxml).length > 0) {
			$main = $('#INTER' + idxml);
			// internal forms
			$('.ParentingClass', $main).css('top', 0 + 'px');

		}

		if ($('#TEMPORAL_aux_' + idxml).length > 0) {
			$main = $('#TEMPORAL_aux_' + idxml);
		}

		var $parent = $main;
		var margin = 18;

		// Get Header
		var idcabecera = undefined;
		idcabecera = $('.ParentingClass[header]', $main).attr('header');


		if (idcabecera == undefined) {
			idcabecera = $($main).attr('header');
			var inner = true;
		}



		// Margin for headers
		if (idcabecera != undefined) {
//            loger('main: '+$main.attr('id')+ ' cab:'+idcabecera);
			cabeceraIng = Histrix.getCustomHeight('DIVFORM' + idcabecera);

			if (cabeceraIng != 0) {
			
			 //loger('main: '+$main.attr('id')+ ' cab:'+idcabecera+ ' h:'+cabeceraIng);
			 
				$parent = $('.ParentingClass', $main).css('top',
						cabeceraIng + 'px');
				if (inner)
					$parent = $($main).css('top', cabeceraIng + 'px');

			}
		} else {

			// HIDES HEADER!!				
	 		$('.ParentingClass', $main).css('top', 0 + 'px');

		}

		// Get table body
		var $tbody = $('#tbody_' + idxml);

		if ($tbody.length > 0) {
			// $tbody.css('height','100%');
			var currentHeight = $tbody.outerHeight();
			// loger('resto');
			var elementsHeight = 0;
			// Calculate Height
			$(
					'#thead_' + idxml + +' , #tfoot_' + idxml + ' , #botonera'
							+ idxml + ' , #totales_' + idxml
							+ ' , #inTableForm' + idxml + ' , #Filtros' + idxml
							+ ' , .paginar' + ' , .form', $parent).each(
					function () {
						elementsHeight += $(this).outerHeight();

					});

			// table headers height
			var elementsHeight2 = 0;
			$( '#thead_' + idxml + +' , #tfoot_' + idxml + ' , #botonera'
							+ idxml + ' , #totales_' + idxml
							+ ' , #inTableForm' + idxml + ' , #Filtros' + idxml
							+ ' , .paginar' + ' , form[tipo=abm-mini]'
							+ ' , form[tipo=ing] .form', $parent).each(
					function () {

						elementsHeight2 += $(this).outerHeight();
					});

			var total 	= $parent.outerHeight() - (elementsHeight + margin + cabeceraIng);
			var tableTotal  = $parent.outerHeight() - (elementsHeight2 + margin * 2);

;
			$tbody.closest('div.tablewrapper').css({
				'height' : tableTotal + "px"
			});

			
			if ($detailmain != null && $detailmain.length > 0) {
				var detailheight = $parent.outerHeight(); // - cabeceraIng;
				$parent.css({
					'height' : detailheight + "px"
				});
			}

			/*
			 * // Apply Css if ((currentHeight > total || cabeceraIng > 0 ) &&
			 * total > 0 ){ $tbody.css({ 'height':total + "px",
			 * 'overflowX':"hidden" }); }
			 */

		} else {

			// tree method
			var filtroHeight = $('.filtro:visible', $parent).outerHeight() + 5;
			$('.contTablaInt.Tree', $parent).css({
				'top' : filtroHeight
			});

		}

		// pdf frame resize
		$(".pdfFrame").each(function () {
			$(this).css("height", $(this).parent().height() - 40);
		});
		
		Histrix.copyTableFooter($tbody.closest('table'));

		
		// resize inner xmls
		$('.ventint table[xml]' ,$main).each(function(){
			var innerxml = $(this).attr('xml');
			if (xml != innerxml){
			    Histrix.calculoAlturas(innerxml);
			}		
		});
		/*
		if ($('#INTER' + idxml).length > 0) {
			$main = $('#INTER' + idxml);
			// internal forms
			$('.ParentingClass', $main).css('top', 0 + 'px');

		} */


		
	},

	/**
	 * Resize all Table Bodies
	 */
	resizeAll : function () {

		$('.slideVert').each(function () {
			Histrix.resizePanel(this.id);

		});

		var tabs = Histrix.tabs[0];
		$('li.activo', tabs).each(function () {
			var xml = this.id.substring(5);
			Histrix.calculoAlturas(xml);
		});

		/*
		 * var fl = tabs.childNodes.length; for (var i=0;i<fl;i++){ var idli =
		 * tabs.childNodes[i].id if(idli){ var xml = idli.substring(5);
		 * Histrix.calculoAlturas(xml); } }
		 */

		Histrix.resizeTabs();

	},
	/**
	 * Custom Method to get Height or
	 * 
	 * @param obj
	 *            element to compute Height
	 * @return object height or 0
	 */
	getCustomHeight : function (obj) {
		var $obj = $('#' + obj);
		var defaultHeight = 0;
		if ($obj)
			defaultHeight = $obj.outerHeight();
		if (defaultHeight == null)
			defaultHeight = 0;
		return defaultHeight;

	},

	/**
	 * Sound Player wrarper
	 * 
	 * @param sound
	 *            audio tag ID
	 */
	playSound : function (sound) {
		// TODO: check if Sound is enabled
		var audio = $('#audio' + sound)[0];
		audio = document.getElementById('audio' + sound);
		if (audio != undefined && typeof audio.play == 'function') {
			audio.play();
		}
	},
	/**
	 * Custom Alert function
	 * 
	 * @param message
	 *            Text message
	 * @param obj
	 *            optional refering object
	 * @param sound
	 *            Optional sound id
	 */
	alerta : function (message, obj, sound, title) {

		if (title == undefined) title = '';
		// prevent empty messages
		if (message == '' || message == undefined) return false;

		if (sound != undefined)
			Histrix.playSound(sound);

		var supra = Histrix.Supracontenido;

		var html = '<div title="'+ title +'"><p>' + message + '</p>';
		html += '<br/>';

	//	var contenido = "printpdf.php?subject="+encodeURIComponent(title)+"&send=true";

		//"&message="+encodeURIComponent(message)+ 
	//	html += '<a onClick="Histrix.loadInnerXML(\'mails_send_xml\', \''; 
	//	html += contenido +'\'  , null, \'' + title + '\', null, null, { width : \'650px\', height : \'90%\', modal : true });cerrarVent(\'Alerta\');">'; 

	//	html += 'Enviar Reporte de Error</a>';
		
		html += '</div>';

		var newdiv = jQuery(html).attr('id', 'Alerta');

		if (obj) {

			var pos = $(obj).offset();
			var tempX = pos.left;
			var tempY = pos.top;
			var h = document.body.offsetHeight;
			var w = document.body.offsetWidth;

			var finalTop = tempY - obj.clientHeight;

			newdiv.css({
				position : 'absolute',
				top : finalTop + 'px',
				left : (tempX + 3) + 'px',
				zIndex: 99999
			})

			newdiv.addClass('minialerta').click(cerrarVent("Alerta"));

			html = '<img src="../img/emblem-important.png" align="middle">'
					+ message;
			$(supra).append(newdiv);
		} else {
			$(function () {
				$(newdiv).dialog({modal:true});
			});
		}
		if (obj)
			setTimeout( function(){ cerrarVent("Alerta");}, 3000);
	},
	/**
	 * PROGRAM LOADERresto
	 * 
	 * @param contenedor id of container object
	 * @param contenido  program to execute
	 * @param options    options
	 * @param title      title
	 * @param menuId     Menu Id
	 * @param loader     php Loader
	 */

	loadXML : function (contenedor, contenido, options, title, menuId, loader,
			reload, menuoptions) {
		var xmldatos = contenedor;
		var idcontenedor = contenedor.replace(".", '_');

		var defaultoptions = {
			helplink : ''
		};

		var loadoptions = $.extend(defaultoptions, menuoptions);

		// Call functions hooked to beforeLoadXml
		jQuery(Histrix.pluginHooks.beforeLoadXml).each(function () {
			this(contenido, {
				programTitle : title,
				container : xmldatos
			});
		});

		if (reload == true) {
			// reload tab to preserv get parameters
			$('#LI' + idcontenedor + ' span').dblclick();
			return true;
		}

		var createDiv = Histrix.creaNuevoDiv(idcontenedor, title, contenido,
				options, menuId, loader, loadoptions); // Create DIV
		Histrix.activartab(idcontenedor); // Enable Tab

		Histrix.buscando = false;
		Histrix.loadingMsg(idcontenedor, title);

		var postOptions = {
			titulo_div : title,
			helplink : loadoptions.helplink,
			'__winid': Histrix.uniqid
		};

		if (menuId) {
			postOptions._menuId = menuId;
		}

		if (loader == undefined)
			loader = '';


		$('#' + idcontenedor).load(loader + contenido, postOptions, function (responseText, textStatus) {
				if (textStatus == 'error'){
					Histrix.loadingMsg(idcontenedor,  responseText, false);
				}

			// Call functions hooked to afterLoadXml
			jQuery(Histrix.pluginHooks.afterLoadXml).each(function () {
				this(contenido, {
					programTitle : title,
					container : xmldatos
				});
			});

			// save post and get data
			$(this).data('htx_get', loader + contenido ).data('htx_post', postOptions);

			Histrix.calculoAlturas(idcontenedor.substring(3));
		});

		return true;

	},

	/**
	 * PROGRAM LOADER in Window
	 * 
	 * @param contenedor   	id of container object
	 * @param contenido 	program to execute
	 * @param options 		GET parameters
	 * @param title 	    title
	 * @param xmlprincipal  mail Xml program
	 * @param idobj 		id
	 * @param winoptions 	window Options
	 */
	loadInnerXML : function (contenedor, contenido, options, title,
			xmlprincipal, idobj, winoptions) {
		// Default windows Options
		var defaultOptions = {
			checkDuplicate : false,
			posX : false,
			posY : false,
			width : false,
			height : false,
			modal : false,
			maximize : false
		}
		var winopt = $.extend(defaultOptions, winoptions);

		contenedor = contenedor.replace(".", '_');
		if (xmlprincipal != null)
			var divxml = xmlprincipal.replace(".", '_');
		// var uid = 'PRN' + uniqid() + contenedor;
		var uid = 'PRN' + contenedor;


		if (winopt.checkDuplicate == true) {
			if ($('#' + uid)[0])
				return false;
		}

		var obj = $('#' + idobj);
		if (obj[0]) {
			options += '&' + idobj + '=' + obj.val();
		}

		var newdiv = jQuery('<div></div>').addClass('ventint').attr("id", uid);
		var internos = $('.ventint');
		var offset = 0;
		if (internos[0]) {
			offset = internos.length * 15;
		}
		if (winopt.width) {
			newdiv.css({
				zindex : '10',
				width : winopt.width,
				height : winopt.height
			});
		}
		newdiv.css({
			position : 'absolute',
			top : 10 + offset + 'px',
			left : 20 + offset + 'px',
			'z-index' : 8
		});
		if (winopt.posY) {
			newdiv.css({
				top : winopt.posY + 'px'
			});
		}
		if (winopt.posX) {
			newdiv.css({
				left : winopt.posX + 'px'
			});
		}

		var barra = barraDrag(uid, title, winopt);
		var barraInf = barraDragInf(uid);
		newdiv.html(barra);

		var contewin = jQuery(
				'<div><div class="esperareloj" style="top:30%;"><b>' + title
						+ '</b><div id="throbber" /></div></div>').attr("id",
				'DIV' + contenedor).addClass('contewin');

		newdiv.append(contewin);
		newdiv.append(barraInf);

		if (winopt.checkDuplicate == true) {
			newdiv.css({
				zindex : '200'
			});
		}

		var supra = $('#' + divxml);
		if (!supra[0]) {
			supra = Histrix.Supracontenido;
		}
		if (!supra[0]) {
			supra = $('#Supracontenido');
		}

		if (winopt.modal) {
			var modal_window = jQuery('<div></div>').attr("id", 'MODAL' + uid)
					.addClass('modalWindow');
			newdiv.css({
				'top' : '10',
				'right' : '0',
				'margin-left' : 'auto',
				'margin-right' : 'auto'
			});

			supra.append(modal_window).append(newdiv);
		} else {
			supra.append(newdiv);
		}
		// $iframe.css('width', modal_window[0].clientWidth +
		// 'px').css('height', modal_window[0].clientHeight);
		// $('#PRN'+contenedor).draggable({ handle: '#dragbarPRN'+contenedor
		// }).resizable();

		$('#' + uid).draggable({
			handle : '#dragbar' + uid,
			containment : 'parent'
		}).resizable();
		/*.touch({
			animate : false,
			sticky : false,
			dragx : true,
			dragy : true,
			rotate : false,
			resort : true,
			scale : false
		});
                               */
                              
//		options['__winid'] = Histrix.uniqid ;

		$(contewin).load(contenido, options, function(){

			$(this).data('htx_get', contenido).data('htx_post', options);			
		});

		$(supra).scrollTo(contewin);
		
		if (winopt.maximize == true)
			Histrix.maximize(uid);

		return true;
	},
	/**
	 * PopUP PROGRAM LOADER
	 * 
	 * @param contenedor
	 *            id of container object
	 * @param contenido
	 *            program to execute
	 */
	loadExternalXML : function (contenedor, contenido) {
		window.open(contenido, contenedor,
				"dependent=yes,resizable=yes,status=yes,toolbar=no,menubar=no");
	},

	/**
	 * Create DIV Container for XML
	 * 
	 * @param nombre Name
	 * @param titulo title
	 * @param contenido xml program
	 * @param opciones Options
	 * @param menuId
	 * @param loader  php loader
	 */
	creaNuevoDiv : function (nombre, titulo, contenido, opciones, menuId,
			loader, menuoptions) {
		var uid = uniqid();
		menuoptions['tabuid'] = uid;

		if ($('#' + nombre).length == 0) {
			Histrix.creaNuevoLi(nombre, titulo, contenido, opciones, menuId,
					loader, menuoptions);
			var supra = Histrix.Supracontenido;
			var newdiv = jQuery('<div ></div>').attr('id', nombre).attr('tabuid', uid).addClass(
					'contenido');
			supra.append(newdiv);
			return true;
		}
		return false;
	},

	/**
	 * Create new TAB Container for XML
	 * 
	 * @param nombre Name
	 * @param titulo title
	 * @param contenido xml program
	 * @param opciones  Options
	 * @param menuId Menu id
	 * @param order order
	 * @param loader php loader
	 */
	creaNuevoLi : function (nombre, titulo, contenido, opciones, menuId, loader,
			menuoptions) {

		$('#LI'+nombre).remove();

		var newli = jQuery('<li></li>').attr('id', 'LI' + nombre).addClass(
				'activo').attr('title', 'doble click para recargar');

		var newspan = jQuery('<span>' + titulo + '</span>').addClass('tabname')
				.click(function () {
					Histrix.activartab(nombre);
				}).dblclick(
						function () {
							var instance = $('[tabuid="' + menuoptions.tabuid +'"] [instance]:first').attr('instance');

							// this BREAKS when launching another pdf while reload
							//$.post("delvars.php?instance=" + instance +'&ln=2145', {'__winid':Histrix.uniqid});

							Histrix.loadXML(nombre, contenido, opciones,
									titulo, menuId, loader, null, menuoptions);
						});

		var closespan = jQuery('<span></span>').addClass('Xcierre').click(
				function () {
					Histrix.closeTab(nombre, menuoptions['tabuid']);
				});
		newli.append(newspan).append(closespan);

		$('.menu_extra', newli).each(
				function (e) {
					var $helpspan = $(this);
					var helppath = $helpspan.attr('hlppath');
					if (helppath) {
						$helpspan.addClass('boton').css({
							color : 'green'
						}).click(
								function (e) {

									Histrix.loadInnerXML('prev4d95d05bf40f5',
											'pdfviewer.php?ro=true&f='
													+ helppath
													+ '&ancho=1024&alto=768',
											null, $helpspan.text(), null, null,
											{
												width : '80%',
												height : '98%',
												modal : true
											});
									Histrix.menu1.hideAll();
									e.stopPropagation();
											
								});
					}
				});
		Histrix.tabs.append(newli).sortable();
		Histrix.resizeTabs();
	},

	resizeTabs : function () {
		// Reset sizes
		$('li ', Histrix.tabs).css({
			width : 'auto'
		});
		$('li .tabname', Histrix.tabs).css({
			width : 'auto'
		});

		var tabbar = Histrix.tabs.outerWidth();
		var tabWidth = 0;
		var count = 0;
		$('li', Histrix.tabs).each(function () {
			tabWidth += $(this).outerWidth();
			count++;
		});
		if (tabWidth > tabbar) {

			var newWidth = (tabbar / count) - 5 * count;
			var spanWidth = newWidth - 15;
			$('li ', Histrix.tabs).css({
				width : newWidth + 'px'
			});
			$('li .tabname', Histrix.tabs).css({
				width : spanWidth + 'px'
			});
		}

	},
	/**
	 * Add row to Table
	 * @param table
	 */
	addRow : function (table) {

		var lastRow = $('tbody tr:last', table);
		var lastRowNumber = parseInt(lastRow.attr('o'));
		if (!isNumber(lastRowNumber)) {
			lastRowNumber = -1;
		}
		var newRowNumber = lastRowNumber + 1;
		var row = jQuery('<tr>').attr('o', newRowNumber);
		// .attr('insert', 'true');
		var xmldatos = $(table).attr('xml');
		var instance = $(table).attr('instance');
		var xmlcontext = $(table).attr('xmlOrig');
		var vars = {
			'addRow' : newRowNumber,
			xmlOrig : xmlcontext,
			'__winid':Histrix.uniqid
		};
		$('tbody', table).append(row);

		row.load("process.php?accion=insert&getFila=true&fila=" + newRowNumber
				+ "&xmldatos=" + xmldatos + "&instance=" + instance, vars,
				function () {
					Histrix.hideMsg();
					// Histrix.registroEventos(table.id);

				});
	},

	/**
	 * Activate Last used TAB
	 * 
	 * @param nombre   name
	 * @param tabsInt  tabList
	 */
	activartab : function (nombre, tabsInt) {
		var tabs = 'tabs';
		if (tabsInt)
			tabs = tabsInt;

		
		var $parent = $('#' + tabs ).closest('div').parent();

		$('#' + tabs + ' li').each(
				function () {
					var $this = $(this);
					$this.removeClass("activo").removeClass("inactivo");

					if (this.id == 'LI' + nombre) {
						$this.addClass("activo");
						$('#' + this.id.substring(2), $parent).css('visibility',
								"visible").css('display', "");

						Histrix.calculoAlturas(nombre.substring(3), true);

					} else {
						$this.addClass("inactivo");
						$('#' + this.id.substring(2), $parent).css('visibility',
								"hidden").css('display', "none");
					}
				});
	},
	/**
	 * Get Last TAB
	 * 
	 * @return ultima Return Last Tab
	 */
	getUltimaTab : function () {
		var ultima = $('#tabs li:last')[0];
		if (ultima)
			return ultima.id.substring(2);
	},

	/**
	 * close Tab
	 * 
	 * @param string  tabname
	 */
	closeTab : function (tabname, uid) {
	    var $tabsToRemove = $('#'+ tabname+ ', #LI'+tabname);


	    

	    var idxml = tabname.substring(3);
	    if (Histrix.gmaps[idxml]){
    		clearInterval(Histrix.gmaps[idxml].interval);
    	    }

	    
//		$('#' + tabname).remove();
//		$('#LI' + tabname).remove();
	    if ($tabsToRemove.length != 0){
	    	var instance = undefined;
			// envio el cierre de session al servidor
			var all = '';
			var tabs = Histrix.tabs[0];
			var fl = tabs.childNodes.length;
			if (fl == 0)
				all = '&all=true';
			var instances = {};
	    	$('[tabuid='+uid+'] [instance]').each(
	    		function(){
	    			instance = $(this).attr('instance');

					instances[instance] = instance;
	    		});
	    	

			$tabsToRemove.remove();
			Histrix.activartab(Histrix.getUltimaTab());

			$.post("delvars.php?"+ all+'&ln=2332', {'__winid':Histrix.uniqid, 'instances': instances});
			
	    }
	},

	/**
	 * toggle side panel
	 */
	/*
	toggleUtilBar : function () {
		var minWidth = 10;
		var maxWidth = 150;
		var currentWidth = 0;
		var src = '';
		var status = 'closed';
		if ($('#utilbar').css('width') == minWidth + 'px') {
			currentWidth = maxWidth;
			src = '../img/adelante.png';
			status = 'open';
		} else {
			currentWidth = minWidth;
			src = '../img/atras.png';
			status = 'closed';
		}
		$('#utilbar').attr('status', status);
		$('.utilbarStatus').attr('src', src);
		$('#utilbar').animate({
			width : currentWidth + "px"
		}, {
			queue : false,
			duration : 400
		});
		// resize screen
		var $width = parseInt($('.Pagina').css('width')) - currentWidth;
		$('#Supracontenido').css({
			'width' : $width + "px"
		});

	},
	 */
	// //////////////////////////////////////
	// CALC METHODS
	// Re calculates the Form
	// OBSOLETE???
	// IN USE???
	// //////////////////////////////////////
	calculo : function (event) {
		Histrix.calculoForm(this.form);
	},
	// //////////////////////////////////////
	// Form recalculate
	// OBSOLETE???
	// IN USE???
	// //////////////////////////////////////
	calculoForm : function (idformu) {
		var $form = $(idformu);
		var elementos = $(':input', $form) // Implemented jQuery power
		elementos.each(function () {
			var elem = $(this);
			if (elem.attr('onformchange')) {
				var formula = elem.attr('onformchange');
				var formulaEvaluada = Histrix.evaluate(formula, $form);
				elem.value = eval(formulaEvaluada);
			}
		});
	},
	// ////////////////////////////
	// Check Form //
	// ////////////////////////////

	checkForm : function (formElement) {
		var formid = formElement.id;
		var result = true;
		var formInstance = $(formElement).attr('instance');
		// validate form elements
		$('input:visible:enabled,textarea,select:visible:enabled', formElement)
				.each(
						function () {
							var $this = $(this);
							// Do not validate Inner forms
							var innerInstance = $this.closest('[instance]').attr('instance');
							var innerform = $this.attr('innerform');
						
							if (innerInstance != formInstance){
							  return;
							}
							
							/* old method
							if (innerform != undefined && innerform != formid) {
							loger('no valido elemento');
							loger(innerform+ '!=' + formid);
								return;
							}
							*/
							
							// check for mandatory fields

							if ($this.attr('required') != ''
									&& $this.attr('required') != undefined) {
								if ($this.val() == '') {
									this.setCustomValidity('');

									this.setCustomValidity('Valor Requerido');
									result = false;
								} else {
									this.setCustomValidity('');

								}
							}
							// chequeo si las ayudas trajeron algo
							if ($this.attr('valid') != undefined) {

								if ($this.attr('valid') == 'false') {
									this.setCustomValidity('');

									this.setCustomValidity('Valor Incorrecto');
									Histrix.alerta('Valor Incorrecto', this);

									$this.addClass('error');
									result = false;
								} else {
									this.setCustomValidity('');

								}
							}
						});

		return result;
	},

	// ////////////////////////////
	// Notification Handling code
	// ////////////////////////////

	notification : function (id, options) {
		if ($('#Notification' + id).length > 0)
			return; // do not replace

		var title = options.title || 'Titulo';
		var content = options.text || 'Texto';
		var fade = options.fade || 0;
		var icon = options.icon || false;
		var notifClass = options.notifClass || 'notification';
		var image = '';
		if (icon)
			image = '<img src="../img/' + icon + '" /> ';

		var divNotification = jQuery('<div ></div>').attr('id',
				'Notification' + id).addClass(notifClass);

		var closeBtn = jQuery('<img src="../img/icon_close.png" />').css({
			'float' : 'right'
		}).click(function () {
			$('#Notification' + id).slideUp('slow', function () {
				$('#Notification' + id).remove()
			});
			options.click();
		});

		var header = jQuery('<div>' + image + title + '</div>').addClass(
				'header');
		var text = jQuery('<p>' + content + '</p>');
		divNotification.append(closeBtn);
		divNotification.append(header);
		divNotification.append(text);

		$('#notifications').append(divNotification);

		if (fade > 0) {
			setTimeout( function () { $('#Notification' + id ).remove();}, fade);
		}

	},

	Postit : function (id, options) {

		if ($('#Postit' + id).length > 0)
			return; // do not replace

		var title = options.title || 'Titulo';
		var content = options.text || 'Texto';
		var fade = options.fade || 0;
		var icon = options.icon || false;
		var notifClass = options.notifClass || 'postit';
		var image = '';
		if (icon)
			image = '<img src="../img/' + icon + '" /> ';

		var lastpostclass = $('.postit').hasClass('foreground'); // get last
																	// postit
																	// position

		var divNotification = jQuery('<div ></div>').attr('id', 'Postit' + id)
				.addClass(notifClass);

		if (lastpostclass)
			divNotification.addClass('foreground');

		// .css({'position':'absolute','right':'200px'});

		var closeBtn = jQuery('<img src="../img/icon_close.png" />').css({
			'float' : 'right'
		});

		closeBtn.click(function () {
			$('#Postit' + id).slideUp('slow', function () {
				$('#Postit' + id).remove()
			});
		});
		if (options.click) {
			closeBtn.click(function () {
				options.click();
			});
		}

		var header = jQuery('<div>' + image + title + '</div>').addClass(
				'header');
		var text = jQuery('<p>' + content + '</p>');
		divNotification.append(closeBtn);

		divNotification.append(header);
		divNotification.append(text);
		divNotification.draggable({
			containment : 'parent'
		}).touch({
			animate : false,
			sticky : false,
			dragx : true,
			dragy : true,
			rotate : false,
			resort : true,
			scale : false
		});
		// $('#postBoard').append(divNotification);

		$('#Supracontenido').append(divNotification);

		if (fade > 0) {
			setTimeout( function() { 
					$("#Postit" + id ).remove();
					}, fade);
		}

	},

	desktopNotify : function (data, id, timeout) {
		if (window.fluid) {
			window.fluid.showGrowlNotification({
				title : data.title,
				description : data.text,
				priority : 1,
				sticky : false,
				identifier : id
			});
		} else if (window.webkitNotifications) {

	
			// Chrome desktop notifications
			if (Histrix.desktopNotifications[id] == undefined) {
				var popup = window.webkitNotifications.createNotification(
						"../img/histrix_blue_button_100.png", data.title,
						data.text);

				Histrix.desktopNotifications[id] = true;

				if (timeout) {
					popup.ondisplay = function () {
						setTimeout(function () {
							popup.cancel();
						}, 3000);
					};
				}

				popup.show();
			}
		} else if (Histrix.Unity){

			// Unity Notifications
			Histrix.Unity.Notification.showNotification(data.title, data.text);		
		}

	},
	/**
	 * Evaluates expresion
	 * 
	 * @param formula formula
	 * @param formId  Form id
	 */
	evaluate : function (formula, form) {
		/*
		// loger('evaluo' + formId);
		var inputElements = $('#' + formId + ' :input'); 	// Implemented
		if ( element || inputElements.length == 0){
		
		    var $base = $(element).closest('form');
		    inputElements = $(':input', $base);

		}
		*/
		inputElements = $(':input', form);	 // jQuery power

		inputElements.each(function (index, elem) {
			if (elem.name != '') {
				var myregexp = new RegExp(elem.name, "g");
				// pongo a cero (ver como hacer con los campos de texo si es que
				// se usa (no creo)
				var valor = elem.value;
				if (valor == '')
					valor = 0;
				formula = formula.replace(myregexp, valor);
			}
		});

		return formula;
	},

	headerOptions : function (xml) {
		var idxml = xml.replace(".", '_');
		var options = '<input class="printOp" type="checkbox" />';
		var jOptions = $('.printOp', '#' + idxml);
		if (jOptions.length == 0) {
			$('.colHeader', '#' + idxml).each(
					function () {
						var value = this.getAttribute('print');
						var name = this.getAttribute('colName');
						var checked = 'checked';
						if (value != 1)
							checked = ' ';  // ' checked="checked" '
						var options = '<input printname="' + name
								+ '" class="printOp" type="checkbox" '
								+ checked + '/>';
						$(this).append(options);
					})
		} else {
			jOptions.toggle();
		}
	},
	/**
	 * PRINTING AND EXPORTING
	 * 
	 * @param idContenido
	 *            id
	 * @param title
	 *            title
	 * @param printFormId
	 *            print form
	 * @param options
	 *            options
	 */
	imprimirpdf : function (idContenido, title, printFormId, options, dir,
			instance) {
		// Get configuration Options
		var printerName = '';
		var orientacion = '';
		var pagesize = '';
		if (printFormId != null) {
			if ($('#' + printFormId)[0] == undefined)
				printFormId = 'ORI' + printFormId;
			printerName = $('#' + printFormId + ' [name=printername]').val();
			orientacion = $('#' + printFormId + ' [name=__orientacion]:checked')
					.val();
			pagesize = $('#' + printFormId + ' [name=__pagesize]').val();
		}
		if (orientacion == undefined)
			orientacion = '';

		var idxml = idContenido.replace(".", '_');
		var fieldOptions = $('.printOp', '#' + idxml);
		var fieldStringOptions = '';

		fieldOptions.each(function () {
			var value = 0;
			if ($(this).is(':checked'))
				value = 1;

			fieldStringOptions += '&__' + this.getAttribute('printname') + '='
					+ value;
		});

		var printer = '&printer=' + printerName;
		var winprop = "resizable=yes,status=no,toolbar=no,menubar=no,location=no";
		switch (options) {
		case 'send':
			options += '&send=true';
			var contenido = "printpdf.php?pdfnom=" + idContenido + '&instance='
					+ instance + '&' + options + "&__orientacion="
					+ orientacion + '&__pagesize=' + pagesize + '&'
					+ fieldStringOptions;

			Histrix.togglePrint(null, 'FormPrint' + idxml);
			Histrix.loadInnerXML('mails_send_xml', contenido, null, title,
					null, null, {
						width : "650px",
						height : "90%",
						modal : true
					});
			break;
		case 'batch':
			window.open("printTandapdf.php?pdfnom=" + idContenido
					+ "&__orientacion=" + orientacion + '&__pagesize='
					+ pagesize + "&dir=" + dir + '&instance=' + instance + '&'
					+ fieldStringOptions, "wintanda", winprop);
			break;
		default:
			window.open("printpdf.php?echo=1&pdfnom=" + idContenido
					+ '&instance=' + instance + '&' + options
					+ "&__orientacion=" + orientacion + '&__pagesize='
					+ pagesize + printer + '&' + fieldStringOptions, "newWin",
					winprop);
			break;
		}
	},
	/**
	 * Show Print Dialog
	 * 
	 * @param btn
	 *            button
	 * @param form
	 *            Print Form
	 */
	showopprint : function (btn, form) {
		var w = $(btn).position();
		$('#ORI' + form).css('left', w.left - 100 + 'px').slideToggle();
	},

	/**
	 * toggle Print Dialog
	 */
	togglePrint : function (btn, form) {
		$('#ORI' + form).slideToggle();
	},

	/**
	 * Impresion de la tanda contenida en la tabla (wrarper function)
	 * 
	 * @param idContenido
	 *            xml id
	 * @param printFormId
	 *            Print Form
	 * @param dir
	 *            string directory
	 */
	imprimirTandapdf : function (idContenido, printFormId, dir, instance) {
		Histrix.imprimirpdf(idContenido, '', printFormId, 'batch', dir, instance);
	},

	/**
	 * Direct Print
	 * 
	 * @param xml
	 *            xml to print
	 * @param directprint
	 *            direct print true/false
	 */
	printExt : function (xml, directprint, triggerObj) {
		var orientacion = 'P';
		var printstr = '';
		if (directprint == 'true') {

			var printername = getURLParameter(xml, 'printer');

			// Get printer name from select box
			if (printername == undefined || printername == '') {
				var xmlform = $(triggerObj).parents('table').attr('xml');
				var idxmlform = xmlform.replace(".", '_');
				var printform = $('#FormPrint' + idxmlform);
				printername = $('[name="printername"]', printform).val();
				printstr = '&printer=' + printername;
			}

			$('#histrixDebug').load(
					xml + '&__orientacion=' + orientacion + printstr, {'__winid':Histrix.uniqid});
		} else {
			Histrix.loadInnerXML('directPrinting' + uniqid(), xml
					+ '&__orientacion=' + orientacion, '', 'Impresion', xml,
					null, {
						modal : true,
						height : '80%'
					});
		}
	},
	/**
	 * ExportFile
	 * 
	 * @param idContenido
	 *            xml file
	 * @param title
	 *            of the file
	 * @param type
	 *            file extension
	 * @param xmlOrig
	 *            xml origin
	 */
	exportFile : function(instance, title, type, xmlOrig, DAT) {
		window.open('process.php?export='+ type +'&titulo=' + title + '&instance='+instance, 'nwods' + '&DAT='+DAT, '');
	},
	// //////////////////////////////////////
	// WINDOW MANAGMENT ROUTINES
	// //////////////////////////////////////
	/**
	 * Maximize
	 * 
	 * @param Elem
	 *            Element
	 * @param btn
	 *            button
	 */
	maximize : function (Elem, btn) {
		var $elem = $('#' + Elem);
		var maxWidth = Histrix.Supracontenido.css('width');
		var maxHeight = Histrix.Supracontenido.css('height');

		$elem.toggleClass('ventint').toggleClass('maximizeVent');

		var position = $elem.position();
		var width = $elem.css('width');
		var height = $elem.css('height');
		var winobject = {
			x : position.left,
			y : position.top,
			w : width,
			h : height
		};

		if (position.left != 0) {
			Histrix.win[Elem] = winobject;
			$elem.css({
				top : '0px',
				left : '0px',
				width : maxWidth - 5,
				height : maxHeight
			});
		} else {
			var winobj = Histrix.win[Elem];
			$elem.css({
				top : winobj.y + 'px',
				left : winobj.x + 'px',
				width : winobj.w,
				height : winobj.h
			});
		}

		Histrix.calculoAlturas(Elem.substring(3), true);
	},

	editRow : function (obj) {

		if (obj.tagName == 'TR') {
			var row = $(obj);
		} else {
			var row = $(obj).parents('tr');
		}
		var instance = $(obj).closest('[instance]').attr('[instance]');
		var numrow = $(row).attr('o');

		var cells = $(row).children();
		var table = $('td', row).parents('table');
		var xmldatos = table.attr('xml');
		// var data = '';
		var myvars = {};
		var colgroupCols = $('col', table);
		var vars = {
			editrow : numrow
		};
		if ($(row).hasClass('editedRow')) {
			$(row)
					.find(
							'input:visible:enabled,textarea,select:visible:enabled')
					.each(function (index, element) {
						var name = $(element).attr('name');
						if (name != undefined)
							myvars[name] = $(element).val();
					});

			myvars['Nro_Fila'] = numrow;

			Histrix.saveRow(row, xmldatos, myvars, numrow);

		} else {
			$(row).addClass('editedRow');
			$('[rel=edit]', row).attr('src', '../img/filesave.png');

			// $(row).load('editRow.php?xmldatos='+xmldatos, vars, Histrix.hideMsg);

			$(colgroupCols).each(
					function (index, element) {
						var cell = $(cells)[index];
						var fieldName = $(element).attr('campo');
						if (fieldName != "Nro_Fila") {
							vars.valor = $(cell).html();
							vars.campo = fieldName;
							vars['__winid'] = Histrix.uniqid;
							$(cell).load(
									"refrescaCampo.php?editrow=true&xmldatos="
											+ xmldatos+ '&instance='+ instance , vars, Histrix.hideMsg)
									.addClass('editable');
						}
					});

		}
	},

	saveRow : function (row, xmldatos, vars, numrow) {
		vars['__winid'] = Histrix.uniqid;
		var instance = $(row).closest('[instance]').attr('[instance]');
		$(row).load(
				"process.php?getFila=true&fila=" + numrow + "&xmldatos="
						+ xmldatos +"&instance=" + instance + "&accion=update", vars, Histrix.hideMsg)
				.removeClass('editedRow');
	},

	clearForm : function (xmldatos, limpiar, btn, options) {
		// Histrix.showMsg('Limpiando');

		var idxmldatos = xmldatos.replace(".", '_');
		var idform = $(btn).attr('idform');

		$('#DIVFORM' + idxmldatos + '.singleForm').each(function () {
			$(this).slideUp();
			return;
		});

		var defaultOptions = {
			filter : true,
			instance:''
		}
		var formOptions = $.extend(defaultOptions, options);

		if (idform == undefined) {
			idform = formOptions.innerForm;
		}

		var form = xmldatos;
		if (xmldatos.search('Form') == -1) {
			form = 'Form' + idxmldatos;
		}
		if (xmldatos.search('FForm') != -1) {
			idxmldatos = idxmldatos.replace('FForm', '');
		} else {
			if (xmldatos.search('Form') != -1) {
				idxmldatos = idxmldatos.replace('Form', '');
			}
		}

		if (formOptions.instance != ''){
			var formu 	 = $('Form[instance="'+ formOptions.instance +'"]')[0];
		} else {
			var formu 	 = $('#' + form)[0];
		}
		
		var xmlOrig  = $(formu).attr('xmlOrig');
		var instance = $(formu).attr('instance');


		var obj;

		if (formu) {
			var fl = formu.length;
			var vars = "";

			// trigger inner cancel
			$('.btn_mini [accion=cancel]',formu).click();
			// re enable key
			$(formu).find('input:disabled.clave').removeAttr('disabled');

			for (i = 0; i < fl; i++) {
				var tempobj = formu.elements[i];
				if (tempobj.type != "button" && 
					tempobj.type != "reset" && 

				   // only  clear input from current forms and not inner ones
				   ( $(tempobj).attr('form') == form ) && ! tempobj.hasAttribute('innerform')

				   ) {

					if ($(tempobj).is('select')) {
						continue;
					} else {
					
						if (tempobj.getAttribute('valauto') != undefined) {
							continue;
						} else {
							if (limpiar == true)
								tempobj.value = '';
						}

						if (tempobj.type == 'checkbox') {
							tempobj.checked = false;

							if (tempobj.getAttribute('default') == '1') {
								tempobj.checked = true;
								tempobj.value = 1;
							}
						}
					}
					// geomap data
					if (tempobj.getAttribute('ctype') != undefined) {
						var strPoints = '';
						var mapId = 'GEOMAP' + $(tempobj).attr('id');
						var map = Histrix.map[mapId];

						if (Histrix.polygonControl[mapId]) {

							var storage = Histrix.polygonControl[mapId].storage[0];

							for ( var i = 0; i < storage.geometry
									.getVertexCount(); i++) {
								var coord = storage.geometry.getVertex(i);
								// strPoints += coord.lat() + ',' + coord.lng()
								// + '|';
							}
						}
						if (Histrix.markerControl[mapId]) {
							// Store 1st Point Only
							var storage = Histrix.markerControl[mapId].storage[0];
							Histrix.markerControl[mapId].storage[0] = null;
							// var coord = storage.geometry.getPoint();
							// Histrix.markerControl[mapId] = [];
							map.clearOverlays(storage);
						}

						valor = strPoints;
					}

					if (tempobj.getAttribute('internal_class')) {
						switch (tempobj.getAttribute('internal_class')) {
						case 'simpleditor':
							$('#' + tempobj.id).wysiwyg('clear');
							break;
						}
					}

				} else {

					if (tempobj.name == 'Grabar') {
						tempobj.innerHTML = '<img src="../img/filesave.png" alt="'
								+ Histrix.i18n['save']
								+ '"  />'
								+ Histrix.i18n['create'] + '&nbsp;';
						tempobj.setAttribute("accion", 'insert');

					}
					if (tempobj.name == 'delete') {
						tempobj.disabled = true;
					}
				}
				vars += tempobj.name + "=" + tempobj.value + '&';
			}

			// Busco las tablas internas en las fichas y las borro
			// (literalmente)
			// obj = formu.getElementsBySelector('[class="sortable"]' ,
			// '[class="sortable resizable"]');

			obj = $('#' + form + ' .sortable').css({'display':'none'});

			obj.each(function (num) {
				var elem = obj[num];
				// Get it's parent
				var pTag = elem.parentNode;
				// Remove
				//pTag.removeChild(elem);
				var pTag2 = pTag.parentNode;
				// pTag2.style.display='none';
				// var obj2 = pTag2.getElementsBySelector('[type="button"]');
				var obj2 = $('[type="button"]', pTag2);
				obj2.each(function () {
					this.disabled = true;
				});

			});
			var delfiltros = false;
			

			//vars['__winid'] = Histrix.uniqid;

			$.post("setData.php?_show=false&modificar=false&delfiltros=" + delfiltros+"&blanqueo=true" + "&instance=" + instance, vars);

			foco(form);

		} else {

			var miobj = $('[innerform="Form' + idform + '"]');
			miobj.each(function (index, elem) {
				if (elem.getAttribute('valauto') != undefined) {
					// continue;
				} else {
					if (limpiar == true)
						elem.value = '';
				}

				if (elem.type == 'checkbox') {
					elem.checked = false;

					if (elem.getAttribute('default') == '1') {
						elem.checked = true;
						elem.value = 1;
					}
				}
			});

			// Busco las tablas internas en las fichas y las borro
			// (literalmente)
			obj = $('#' + idxmldatos + ' [form="' + idxmldatos + '"]');
			obj.each(function (index, elem) {
				if (limpiar == true) {
					elem.value = '';
				}
			});

			obj = $('#' + idxmldatos + ' [type="button"]');
			obj.each(function (index, elem) {
				if (elem.name == 'Grabar') {
					$(elem).html(
							'<img src="../img/filesave.png" alt="'
									+ Histrix.i18n['save'] + '"  />'
									+ Histrix.i18n['create'] + '&nbsp;').attr(
							"accion", 'insert');
				}
				if (elem.name == 'delete') {
					elem.disabled = true;
				}
			});
		}
		// Histrix.hideMsg();
	},

	geoDrawPoint : function (mapId, pointin) {
		var map = Histrix.map[mapId];
		var bounds = new GLatLngBounds();

		var point = pointin.split('|');
		var len = point.length;
		var pArray = [];
		for ( var i = 0; i < len; i++) {
			var myPoint = point[i].split(',');
			if (myPoint.length > 1) {
				pArray[i] = new GLatLng(myPoint[0], myPoint[1]);
				bounds.extend(pArray[i]);
			}
		}

		var marker = Histrix.markerControl[mapId].createMarker(pArray[0]);

		map.clearOverlays(marker);
		map.addOverlay(marker);

		// map.setZoom(map.getBoundsZoomLevel(bounds) - 1 );
		map.setCenter(bounds.getCenter());

	},
	geoDrawPolygon : function (mapId, polygon) {

		var map = Histrix.map[mapId];
		var bounds = new GLatLngBounds();

		var points = polygon.split('|');
		var len = points.length;
		var pArray = [];
		for ( var i = 0; i < len; i++) {
			var point = points[i].split(',');
			if (point.length > 1) {
				pArray[i] = new GLatLng(point[0], point[1]);
				bounds.extend(pArray[i]);
			}

		}
		var polygon = Histrix.polygonControl[mapId].createPolygon(pArray);

		map.clearOverlays(polygon);
		map.addOverlay(polygon);
		map.setZoom(map.getBoundsZoomLevel(bounds) - 1);
		map.setCenter(bounds.getCenter());

	},

	// get Cell value
	// if there is an input will take its value

	getCellValue : function (cell) {
		var valor = 0.00;
		$cell= $(cell);
		if (cell.lastChild) {
			var inp = cell.lastChild;

			if (inp.tagName == 'INPUT') {
				valor = $(cell.lastChild).val();
			}
			else {
				if (cell.tagName == 'TD') {
					valor = $cell.text();
					if (cell.hasAttribute('valor'))
						valor = $cell.attr('valor');
				}
			}				
			
		} else {
			if (cell.tagName == 'TD') {

				valor = $cell.text();
				if (cell.hasAttribute('valor'))
					valor = $cell.attr('valor');

			}
		}
		
		// valor = parseFloat(valor);
		
		return valor;
	},

    getSelectedRow: function (idxml) {
		var myRow = $('#tbody_' + idxml + ' .trsel');
		return myRow;
	},

    selectRow: function (tr) {
		var $tr = $(tr);
		$('.trsel', $(tr.parentNode)).removeClass('trsel');
		$tr.addClass('trsel');
		return $tr;
	},

    getRowValues: function(tr, form, className) {

		var cols = tr.getElementsByTagName("td");
		var columnas = cols.length;
		var vars;
		var parentTable = getParent(tr, 'TABLE');
		var colgroupCols = $('col', parentTable);

		if (className) {
			$('.' + className, tr).each(
					function () {
						$input = $('input ', this);
						if ($input.length > 0) {
							// loger($input);

							vars += '&' + $(this).attr('campo') + '='
									+ $input.prop('defaultValue');
							$input.prop('defaultValue', $input.val());
						} else {
							var value = $(this).html();
							if ($(this).attr('valauto') != undefined)
								value = $(this).attr('valauto');

							if ($(this).attr('valor') != undefined)
								value = $(this).attr('valor');
							vars += '&' + $(this).attr('campo') + '=' + value;
						}
					});
			return vars;
		}

		for (i = 0; i < columnas; i++) {
			if (colgroupCols[i]) {

				var destino = colgroupCols[i].getAttribute('campo')
						|| cols[i].getAttribute('campo');

				if (cols[i].getAttribute('valauto') != undefined) {
					valor = cols[i].getAttribute('valauto');

				}

				if (cols[i].getAttribute('valor') != undefined) {
					valor = cols[i].getAttribute('valor');
				} else {
					var inputField = $('input[name="' + destino
							+ '"], select[name="' + destino + '"]', tr);

					if (inputField.length > 0) {
						valor = inputField.val();
						var tempobj = inputField[0];
						if (tempobj.type == "checkbox") {
							if (tempobj.checked)
								valor = 1;
							else
								valor = 0;
						}
						if (tempobj.type == "radio") {
							if (tempobj.checked) {
								valor = tempobj.value;
							} else
								continue;
						}

					} else {
						valor = $(cols[i]).html();
						if (valor == null)
							valor = cols[i].innerText;
					}
				}
				if ($(form)) {
					var input = $(destino);
					if (valor != '') {
						input.value = valor;
					}
				}
				vars += '&' + destino + '=' + valor;
			}
		}
		return vars;
	},

	// get position of elemet from event
	getPosition : function (e) {
		e = e || window.event;
		var cursor = {
			x : 0,
			y : 0
		};
		if (e.pageX || e.pageY) {
			cursor.x = e.pageX;
			cursor.y = e.pageY;
		} else {
			var de = document.documentElement;
			var b = document.body;
			cursor.x = e.clientX + (de.scrollLeft || b.scrollLeft)
					- (de.clientLeft || 0);
			cursor.y = e.clientY + (de.scrollTop || b.scrollTop)
					- (de.clientTop || 0);
		}
		return cursor;
	},

	// Calculo el total de un campo en la grilla
	calculoTotal : function (campo, uid, act, postoptions) {
	//	loger('jQuery calculoTotal ' + campo + ' uid' + uid + ' act:' + act);

		if (uid != undefined && uid != null) {
			var $tempfield = $('[uid=' + uid + ']');

			var tempfield = $tempfield[0];
			if (tempfield != undefined)
				campo = tempfield;

		}
		var suma = 0.00;

		var defaultOptions = {
			totalvalue : null
		}
		var options = $.extend(defaultOptions, postoptions);

	//	loger(options);
		//
		// Explicit Value
		if (options.totalvalue != null && options.totalvalue != '') {
			var instance = options.instance;
			var $form = $('form[instance="' + instance + '"]');

			suma = options.totalvalue;
			campo = $('[name="' + options.field + '"]', $form)[0];

		} else {

			// calculate Value
			//loger('campo' + campo);
			 
			var cell = $(options.cell).get(0);

			if (campo == undefined && cell == undefined)
				return;

			if (cell != undefined)
			    var celda = cell;
			else 
			    var celda = campo.parentNode;

			if (celda) {




				var $form = $(celda).closest('form');
				var cellindex = celda.cellIndex;	

				var tabla = $(celda).closest('tbody')[0];
				var nr = tabla.rows.length;
				var valor = 0.00;


				for (i = 0; i < nr; i++) {
					var fila = tabla.rows[i];
					var celdaaSumar = fila.cells[cellindex];

					if (celdaaSumar) {
						
						valor = Histrix.getCellValue(celdaaSumar);
						
						if (Histrix.isTime(valor)){
							
							if (suma == 0.00)
								suma = '00:00:00';

							suma = Histrix.sumTime(suma, valor);
						} else {
    							valor = parseFloat(valor);
						    
							if (!isNaN(valor)){
								suma += valor;
							}
						}
					}
				}
				

			} else {
				// loger('no hay celda');
			}
		}


		// loger($('#total_'+campo.name));
		// loger(suma);
	//	if (suma)
	//		suma = suma.replace(/,/, '');

		var importe = suma;
		try {
			importe = suma.toFixed(2);
		} catch (ex) {

		}
		importe = importe.replace(/,/, '');
                if (importe == '0.00'){
                	importe = 0;
                }
		var evaldest = undefined;
		if (campo != undefined) {

			var destino = $('#total_' + campo.name, $form).html(importe);

			//var destsTotal = 'evaldest= ' + destino.attr('jsevaldest') + ';';
			//eval(destsTotal);
			
			var evaldest = jQuery.parseJSON(destino.attr('jsevaldest'));

			var parentInstance = destino.attr('jsparent');
		}
		else {

			var $footercell = $(tabla).closest('table').children('tfoot').children().children().eq(cellindex );

			$footercell.css('color', 'red');
			$footercell.html(importe);

			$('#cloned_' + $footercell.attr('id')).html(importe);

		}

		$parentInstance = $(campo).closest(
				'[instance="' + parentInstance + '"]');

		if (evaldest != undefined)
			for ( var k = 0; k < evaldest.length; k++) {
				// if(evaldest){
				var formid = $(campo).closest('form').attr('id');
				var destino2 = $('#' + formid + ' [name="' + evaldest[k] + '"]');

				// loger('#' + formid + ' [name="' + evaldest +'"]');
				if (destino2[0] == undefined) {
					destino2 = $('[name="' + evaldest[k] + '"]',
							$parentInstance);
				}

				if (destino2[0] == undefined) {
					destino2 = $('[name="' + evaldest[k] + '"]');
				}

				if (destino2[0] == undefined)
					continue;

				var dest2 = destino2[0];

				if (dest2.tagName == 'INPUT') {

					dest2.value = importe;
					$(dest2).change();

					var xmldatos = $(dest2.form).attr('xml');
					// seteo el valor del campo en el contenedor
					// comentado NO ANDA la suma de las grillas.

					setCampo(evaldest[k], dest2.value, xmldatos, 'false', dest2);

				}

				if (dest2.getAttribute('jsevaldest') != undefined) {

					//var dests = 'var array_destinos= ' + dest2.getAttribute('jsevaldest');
					//eval (dests);

					var array_destinos = jQuery.parseJSON(dest2.getAttribute('jsevaldest'));
					var calc  = 'var array_calcular= ' + dest2.getAttribute('jseval');

					eval(calc);



					for ( var j = 0; j < array_destinos.length; j++) {
						var jseval = array_calcular[j];
						var jsevaldest = array_destinos[j];

						calculojs(jseval, jsevaldest, dest2.form.id, dest2,
								null, false, 'false');
					}
				}

				if (campo) {
					if (act == true) {
						var xmlParent = $(campo.form).attr('xml');
						if (tabla != undefined) {
							var miJSON = Histrix.updateGrid(tabla);
							var grilla = tabla.parentNode.parentNode.id; // rehacer
							var instance = $(tabla).closest('table').attr(
									'instance');
						}
						if (instance != undefined && grilla != undefined) {
							$.post("setData.php?_show=false"
									+ '&instance=' + instance + '&xmlOrig='
									+ xmlParent + "&actualizoTabla=true", {
								mijson : miJSON,
								'__winid':Histrix.uniqid
							});
						}
					}
				}
			}

	},

	isTime: function isTime(timeString) {
		var objRegExp  = /(^\d{2}:\d{2}:\d{2}$)/;
		return objRegExp.test(timeString);
		/*
		if (!objRegExp.test(timeString))
			return false;
		else 
			return true;*/
	},
	sumTime: function sumTime(time1, time2) {

		var hours = parseInt(time1.substr(0,2)) + parseInt(time2.substr(0,2));
		var mins  = parseInt(time1.substr(3,2)) + parseInt(time2.substr(3,2));
		var secs  = parseInt(time1.substr(6,2)) + parseInt(time2.substr(6,2));
		
		var sum = hours * 3600 + mins * 60 + secs ;		

		var hoursSum = Math.floor( sum / 3600);
		var hSum = hoursSum ;		

		if (hoursSum < 9){
			hSum = '0' + hoursSum ;
		}

		var hMinutes = Math.floor( (sum -  (hoursSum * 3600) ) / 60);

		var hMin = hMinutes ;		
		if (hMinutes < 9){
			hMin = '0' + hMinutes ;
		}

		var hSeconds = sum  -  (hoursSum * 3600) - (hMinutes * 60);
		var hSecs = hSeconds;
		if (hSeconds < 9){
			hSecs = '0' + hSeconds ;
		}

		var sumString = hSum + ':' + hMin + ':' + hSecs;


		return sumString;
	},	
	cargoDetalle : function (destino, Fila, Midiv, options) {
		// loger('cargoDetalle jQuery');

		var inline = options.inline || false;

		var variable = {
			__inline : false,
			__inlineid : Midiv
		};
		// var e = window.event;
		var tabla = $(Fila).closest('table');
		// Fila.parentNode.parentNode;

		var Row = $(Fila);
		if (inline) {
			var cell = $(Fila);
			Row = $(Fila).parent('tr');
			tabla = $(Row).closest('table');

			if (Row.attr('openrow') == 'true') {

				Row.attr('openrow', 'false');

				$('span', cell).removeClass('ui-icon-triangle-1-s').addClass(
						'ui-icon-triangle-1-e');

				tabla[0].deleteRow($(Row)[0].rowIndex + 1);
				return;
			} else {
				Row.attr('openrow', 'true');

				// removeClass('inline').addClass('inlineOn');
				$('span', cell).removeClass('ui-icon-triangle-1-e').addClass(
						'ui-icon-triangle-1-s');

				var cols = $(Row)[0].getElementsByTagName("td");
				var length = cols.length;
				var uid = uniqid();
				Midiv = Midiv + uid;
				var container = '<tr ><td class="detailTD"  colspan="' + length
						+ '"><div  id="' + Midiv
						+ '" >Cargando...</div></td></tr>';
				$(Row).after(container);
				var variable = {
					__inline : true,
					__inlineid : Midiv
				};
			}

		} else {
			$('.trsel', tabla).removeClass('trsel');
			$(Fila).addClass('trsel');

			var oldInstance = $('#' + Midiv+' [instance]:first').attr('instance');

			Histrix.loadingMsg(Midiv, 'Cargando...');
		}
		variable['__winid'] = Histrix.uniqid;

		// Clean previous Containers
		// 

		$.post("delvars.php?instance=" + oldInstance +'&ln=3666', {'__winid':Histrix.uniqid});

		$('#' + Midiv).load(destino, variable, function () {
			// Callback function
			Histrix.calculoAlturas(Midiv);

			//save get and post data
			$(this).data('htx_get', destino).data('htx_post', variable);

		});
	},

	reloadXml: function reloadXml(container){

		var getData  = $(container).data('htx_get');
		var postData = $(container).data('htx_post');
		postData['__winid']=Histrix.uniqid;

		$(container).load(getData, postData, function(){

			// after reload
		})
	},

	/*
	 * Genero una Ventana interna leyendo un nuevo xml cuyo resultado va a parar al
	 * padre
	 */
	ventInt:function ventInt(xmlpadre, xmlhijo, parametros, titulo, winoptions) {
		// var myregexp = new RegExp('.xml',"g");

		var defaultOptions = {
			modal : false
		}
		var winopt = $.extend(defaultOptions, winoptions);

		var idxmlpadre = xmlpadre.replace(".", '_');
		var idxmlhijo = xmlhijo.replace(".", '_');

		var tempX = 10;
		var tempY = 10;
		// var titulo = '';
		var offset = 0;
		var internos = $('.ventint');
		if (internos) {
			offset = internos.length * 15;
		}

		if ($('#DIV' + idxmlhijo)[0] == undefined) {

			var supra = $('[instance="' + winoptions.parentInstance+ '"]:first');

			if (winopt.modal) {
				supra = supra.closest('.detalle');
				if (supra.length == 0)
					supra = $('[instance="' + winoptions.parentInstance+ '"]').closest('.contenido');
			}
			/*
			if (supra[0] == undefined)
				supra = $('#' + idxmlpadre);
			*/
		
			// whe is an inner window
			if (supra[0] == undefined) {
				supra = $('#INT' + idxmlpadre).closest('.detalle');
			}

			// fall back
			if (supra[0] == undefined) {
				supra = $('#Supracontenido');
			}


			var newdiv = jQuery('<div></div>').attr("id", 'DIV' + idxmlhijo)
					.addClass('ventint').css({
						top : (tempY + offset) + 'px',
						left : (tempX + offset) + 'px'
					});

			if (winopt.modal) {
				var modal_window = jQuery('<div></div>').attr("id",
						'MODALDIV' + idxmlhijo).addClass('modalWindow');
				newdiv.css({
					'top' : '0',
					'right' : '0',
					'margin-left' : 'auto',
					'margin-right' : 'auto',
					'width':winopt.width
				});
				supra.append(modal_window).append(newdiv);
			} else {
				supra.append(newdiv);
			}

			// newdiv.style.Height = 'auto';
			// var contenido =
			// 'AbmGenerico.php?xmlOrig='+xmlpadre+'&xml='+xmlhijo+parametros;
			var contenido = 'histrixLoader.php?xmlOrig=' + xmlpadre + '&xml='
					+ xmlhijo + parametros;
			var dragbar = barraDrag('DIV' + idxmlhijo, titulo);
			var divresize = barrawin('DIV' + idxmlhijo);

			var newdiv2 = jQuery('<div></div>').addClass('contewin').attr('id',
					'INTER' + idxmlhijo);

			Histrix.loadingMsg(newdiv2[0].id, titulo);
			newdiv2.load(contenido, {'__winid':Histrix.uniqid, 'parentInstance':winoptions.parentInstance}, function () {
				foco('Form' + idxmlhijo);
				Histrix.calculoAlturas(xmlhijo);
			});

			newdiv.append(dragbar).append(newdiv2).draggable({
				handle : '#dragbarDIV' + idxmlhijo
			});
			// .append(divresize);
		}
		// low other windows
		$('.ui-draggable', supra).css('z-index', '1');

	},
	/**
	 * Activate multiple input boxes
	 */
	activateCheck : function (campo, aActivar, desactivar) {
	    	var form = $(campo).closest('form');

		if (campo.type == 'checkbox') {
			for ( var i = 0; i < aActivar.length; i++) {
				$('[name=' + aActivar[i] + ']', form).each(function () {

					var idlbl = this.getAttribute('idlbl');
					$lbl = $(idlbl);
					if (campo.checked) {

				//		this.disabled = state;
//						this.readOnly = state;
						if (desactivar){
						    $(this).attr('readonly', 'readonly').attr('disabled', 'disabled');
						       $lbl.attr('readonly', 'readonly').attr('disabled', 'disabled');
							if ($(this).is('div'))
								$(this).css({
									'display' : 'none'
								});

						} else {
							 $(this).removeAttr('disabled').removeAttr('readonly');
							 $lbl.removeAttr('disabled').removeAttr('readonly');
							if ($(this).is('div'))
								$(this).css({
									'display' : 'block'
								});

						}
                                                             /*
						if ($lbl){
							$lbl.disabled = desactivar;
							$lbl.readOnly = desactivar;
						}

						if ($(this).is('div'))
							$(this).css({
								'display' : 'none'
							});
                            			*/
					} else {

						if (desactivar){
							$(this).removeAttr('disabled').removeAttr('readonly');
							$lbl.removeAttr('disabled').removeAttr('readonly');
							if ($(this).is('div'))
								$(this).css({
									'display' : 'block'
								});

						} else {
						    $(this).attr('readonly', 'readonly').attr('disabled', 'disabled');
						    $lbl.attr('readonly', 'readonly').attr('disabled', 'disabled');
						
							if ($(this).is('div'))
								$(this).css({
									'display' : 'none'
								});

						}
                                                            /*
//						this.readOnly = !state;
						if ($lbl){
							$lbl.disabled = !desactivar;
							$lbl.readOnly = !desactivar;
						}
						
						if ($(this).is('div'))
							$(this).css({
								'display' : 'block'
							});
							*/
					}
					$("input.date:enabled").datepicker();

				});
			}
		}
	},

	/**
	 * Increase / Decrease font Size
	 */
	changeFont : function (f) {
		document.body.style.fontSize = f + 'px';
		$('.XulMenu').css('fontSize', f + 'px');
	},

	/**
	 * toggle all Check boxes in a table
	 */
	checkToggle : function (element, event) {
		var checked = $(element).is(':checked');

		var th = $(element).closest('th');
		var table = $(element).closest('TABLE');
		var grilla = $(table).attr('xml');
		var name = $(th).attr('colname');
		var fields = [];
		var refresh = false;
		$(' [campo="' + name + '"] > input:enabled', $(table)).each(
				function (index, input) {
					$input = $(input);
					// if column check is true
					// check the checkbox
		
					if (checked){

					    // if current check is not  checked
					    if (!$input.is(':checked')){
						$input.attr('checked', checked).prop('checked', checked);

						if ($input.attr('refresh') == 'true') {
							$input.attr('checked', true).click();
							refresh = true;
						} else {
							$input.change();

						}


					    }
					} else {
					    // ifcurent check is  checked
					    if ($input.is(':checked')){
						$input.attr('checked', checked).prop('checked', checked);

						if ($input.attr('refresh') == 'true') {
							$input.attr('checked',false).click();

							refresh = true;
						} else {
							$input.change();

						}


					    }

					}

					 //$input[0].checked = check;
					 
				});

		// var xmlParent = $(campo.form).attr('xml');
		if (table[0] != undefined && refresh == false) {
			var miJSON = Histrix.updateGrid(table[0], name);

			var instance = $(table).closest('table').attr('instance');

			// loger('Set Table Data');
			$.post("setData.php?_show=false&instance="
					+ instance + "&actualizoTabla=true", {
				mijson : miJSON,
				'__winid':Histrix.uniqid
			});
		}

	},

	/**
	 * Show messages
	 * 
	 * @param text
	 *            Text
	 */
	showMsg: function(text, modal) {
		
		if (modal == true) {
			var $msg  =jQuery('<div><p><span class="ui-icon ui-icon-alert" style="float: left; margin: 0 7px 20px 0;"></span>'+text+'</p></div>').dialog(
			    {modal:true,
			    resizable: false,
				draggable: false,
				height:80
			    });

			$('.ui-dialog-titlebar',$msg.parent()).css({'display':'none'});

			return $msg;
		} 
		mensaje = Histrix.Msg;
		mensaje.html( text + '<div style="float:right;margin:2px;" id="throbber" />').css("visibility", "visible");

	},

	/**
	 * hide Messages
	 */
	hideMsg:function () {
		mensaje = Histrix.Msg;
		mensaje.css("visibility", 'hidden');
		$('#MODALMSG').remove();
	},	

	updateGridById: function(grilla) {

		var idgrilla = grilla.replace(".", '_');
		var tabla = $('#tbody_' + idgrilla)[0];
		return Histrix.updateGrid(tabla);
	},

	/**
	 * Update Datacontainer with screen grid data
	 */
	updateGrid : function (tabla, field) {
		var nr = tabla.rows.length;
		var tablaDatos = [];
		var tablaCompleta = [];
		var titulosColumnas = [];

		var nombre = '';
		var rowcount = 0;

		for (i = 0; i < nr; i++) {

			var fila = tabla.rows[i];
			var elems = fila.childNodes.length;
			var valor = '';
			var valoresFilas = [];
			var col = 0;
			for (j = 0; j < elems; j++) {

				valor = null;
				var celda = fila.childNodes.item(j);

				nombre = celda.getAttribute('campo');

				if (celda.getAttribute('valor') != undefined) {
					valor = celda.getAttribute('valor');
				} else {
					valor = $(celda).text();
				}
				// Para los campos con Inputs
				
				if (celda.firstChild) {
					var inp = celda.firstChild;
					// var nombre= inp.name;
                                            /*
					if (inp.tagName == 'SELECT') {
						valor = $(inp).val();
					}
					*/
					if (inp.tagName == 'INPUT') {
						valor = $(inp).val();
						if (inp.type == "checkbox") {
							if (inp.checked)
								valor = 1;
							else
								valor = 0;
						}
					}
					
					if (inp.tagName == 'BUTTON') {
						valor = $(inp).val();
					}
				}
				
				
				$(':input[name]:first', celda).each(function (){
					valor = $(this).val();
					
					if (this.type == "radio") {
					       valor = $("input:checked", celda).val();
					}
					
					
					// TO FIX SOME Chckboxes do not have name
					if (this.type == "checkbox") {
						if (this.checked)
							valor = 1;
						else
							valor = 0;
					}
				});
				
				
				if (field == undefined) {
					if (nombre != undefined) {
						titulosColumnas[j] = nombre;
					}

					valoresFilas[j] = valor;
				} else {
					if (nombre == field && valor != undefined) {
						titulosColumnas[col] = nombre;
						valoresFilas[col] = valor;
						col++;

					}
				}
			}
			if (valoresFilas.length != 0) {
				tablaDatos[rowcount] = valoresFilas;
				rowcount++;
			}
		}
		tablaCompleta[0] = titulosColumnas;
		tablaCompleta[1] = tablaDatos;

		// var miJSON= Object.toJSON(tablaCompleta);
		var miJSON = $.toJSON(tablaCompleta);
		return miJSON;
	},
	
	getNodeData: function(Node){
	    var data = {};
	    $(':input', Node).each(function(){
		data[this.name]=encodeURIComponent($(this).val());
	    });
	    return data;
	}


}

// TODO: move this functions into Histrix Object.

/** INICIO */
//
// String Functions
// Add left trim, right trim, and trim functions
if (!String.prototype.lTrim) {
	String.prototype.lTrim = function () {
		return this.replace(/^\s*/, '');
	}
}
if (!String.prototype.rTrim) {
	String.prototype.rTrim = function () {
		return this.replace(/\s*$/, '');
	}
}
if (!String.prototype.trim) {
	String.prototype.trim = function () {
		return this.lTrim().rTrim();
	}
}

function trimNumber(s, caracter) {
	if (caracter == undefined)
		caracter = '0';
	while (s.substr(0, 1) == caracter && s.length > 1) {
		s = s.substr(1, 9999);
	}
	return s;
}

function pad(number, length, character) {

	var str = '' + number;
	while (str.length < length) {
		str = '' + character + '' + str;
	}

	return str;

}

function jsextract(input, destinos, posini, posfin) {
	// loger('jsextract jQuery');
	var formActual = input.form.id;

	// Opera hack
	if (formActual == null) {
		formActual = input.getAttribute('form');
	}

	var valor = input.value;
	for ( var i = 0; i < destinos.length; i++) {
		$('[name=' + destinos[i] + '][form=' + formActual + ']').each(
				function () {
					var sub = valor.substring(posini[i], posfin[i]);
					this.value = sub;
					if (this.getAttribute('onchange') != undefined) {
						var onc1 = this.getAttribute('onchange')

						if (onc1.substring(0, 6) == 'buscar'
								|| onc1.search('setCampoSilent') != -1) {
							$(this).change();
						}
					}
				});
	}
}

function isNumber(val) {
	return /^-?((\d+\.?\d?)|(\.\d+))$/.test(val);
}

function calculojs2(obj) {
	// loger(obj);
	loger('Calculando Version 2' + obj.id);

	if (obj.getAttribute('jsevaldest') != undefined) {

		//var dests = 'var array_destinos= ' + obj.getAttribute('jsevaldest');
		//eval (dests);

		var array_destinos = jQuery.parseJSON(obj.getAttribute('jsevaldest'));

		var calc = 'var array_calcular= ' + obj.getAttribute('jseval');
		eval(calc);

		var formuTemp = obj.form.id;

		for ( var j = 0; j < array_destinos.length; j++) {
			var formula = array_calcular[j];
			var jsevaldest = array_destinos[j];

			var objDestino = $('#' + formuTemp + ' [name="' + jsevaldest + '"]'); // DESTINO

			var formulaArray = formula.split(' ');
			for ( var fi = 0; fi < formulaArray.length; fi++) {
				if (formulaArray[fi].length > 1) {

					// tempobj =
					// formu.select('[name="'+formulaArray[fi]+'"]')[0];
					tempobj = $('#' + formuTemp + ' [name="' + formulaArray[fi]
							+ '"]')[0];
					if (tempobj) {
						var valor = tempobj.value;
						if (valor == '')
							valor = 0;
						if (valor == undefined)
							valor = 0;

						formulaArray[fi] = valor;
					}
				}
				// TEST is this necesary?
				if (formulaArray[fi] == " ' ") {
					formulaArray[fi] = "'";
				}
			}
			formula = formulaArray.join("");

			eval('var valor= ' + formula + ';');
			objDestino.val(valor);

		}
	}
}

function calculojs(formula, destino, formulario, obj, formExtra, actxml) {

	loger('formulario'+formulario+'obj'+obj+'calculojs:'+formula+' destino:'+destino);
	var $tempobj;
	var tempobj;
	var xmldatos;
	var formuTemp = formulario;

	if (obj.form) {
		// loger(obj.form);
		formuTemp = $(obj).closest('form').attr('id');
		// loger('id');
		// loger(formuTemp);

	}
	if ($(obj).attr('innerform') != undefined && $('#' + formuTemp).length == 0) {
		formuTemp = 'TR' + $(obj).attr('innerform');
	}

//	loger('formulario' + formulario + 'obj' + obj + 'calculojs:' + formula +  ' destino:' + destino);

	// for grid calculations
//	var objtr = getParent(obj, 'TR');
	var  $formu  = '';

	var $row = $(obj).closest('tr');;
	var rowNumber = $row.attr('o');
	if (rowNumber >= 0) {
		var $formu = $(obj).closest('form');
		
	}

	
  	formuTemp = formuTemp.replace(".", '_');
	

	if ($formu.length == 0){
		var $formu = $(obj).closest('[instance]') ;
	}

	// fixes form names with dots.
	if ($formu.length == 0){
	    var  $formu  =  $row.closest('div.contTablaInt');
	}
	

	var formu = $formu[0];

	var valor = 0;

	// get extra or parent form
	if (formExtra != null && formExtra != '') {
		//alert(formExtra);
		var idformExtra = formExtra.replace(".", '_');
		
	}
//	loger(formuTemp);
//	loger(formu);
	if (formu) {
		// SECOND METHOD JUST SEARCH FORMULA
		var calls = 0;
		var formulaArray = formula.split(' ');
		for ( var fi = 0; fi < formulaArray.length; fi++) {
			if (formulaArray[fi].length > 1) {
				// grid calculation
				if (rowNumber >= 0) {
					// cell values
					$tempobj = $('input[name="' + formulaArray[fi] + '"]', $row);

					if ($tempobj.is('input')) {
						valor = $tempobj.val();
					} else {
						$tempobj = $('[campo="' + formulaArray[fi] + '"]', $row);
						if ($tempobj.length) {
							valor = $tempobj.html();
						}
					}

					tempobj = $tempobj[0];
					if (tempobj != undefined){
						if (tempobj.type == "checkbox") {
							if (tempobj.checked)
								valor = 1;
							else
								valor = 0;

						}
					}

				} else {

					// search value in current form
					$tempobj = $('[name="' + formulaArray[fi] + '"]', $formu);

					// search in parent container
					if ($tempobj.length == 0)
						$tempobj = $('[name="' + formulaArray[fi] + '"]', $('#' + idformExtra));

					if ($tempobj.length) {
						valor = $tempobj.val();

						tempobj = $tempobj[0];
						if (tempobj.type == "checkbox") {
							if (tempobj.checked)
								valor = 1;
							else
								valor = 0;

						}
						if (tempobj.type == "radio") {
							if (tempobj.checked) {
								valor = tempobj.value;
							} else
								continue;
						}
					}

				}

				if ($tempobj.length) {
					if (valor == '')
						valor = 0;
					if (valor == undefined)
						valor = 0;

					formulaArray[fi] = valor;
				}
			}
		}
		formula = formulaArray.join("");
		try {

			eval('valor= ' + formula + ';');
//			loger(formula);

		} catch (ex) {
			valor = false;
			loger('eval exception:' + formula + ';');
			return false;
		}


		if (destino == '__EVAL') {
			$buttons = $('[type="button"]', $formu);

			//$currentTableFormButton = $( 'button [name="Cargar"]' ,$(obj).closest('table .contTablaInt'));

			if (valor == false) {

				var errorMessage = $(obj).attr('errorMessage')
						|| 'Valor Incorrecto..';
				loger(errorMessage + ': ' + formula);
				Histrix.alerta(errorMessage, obj);
				$(obj).attr('valid','false');
				setTimeout( function() { 
								$("#" + obj.id ).focus();
							}, 1);

				// disable currentForm buttons

			//	$currentTableFormButton.attr('disabled', 'disabled');

				// disable Form
				$buttons.each(function (num) {
					var elem = $buttons[num];
					if (elem.name == 'Grabar' || elem.name == 'Insert') {
						elem.setAttribute('disabled', 'disabled');
					}
				});
				return false;
			} else {
				// re-enable Form
			//	$currentTableFormButton.removeAttr('disabled');
				$(obj).attr('valid','true');

				$buttons.each(function (num) {
					var elem = $buttons[num];
					if (elem.name == 'Grabar' || elem.name == 'Insert') {
						elem.removeAttribute('disabled');
					}
				});
			}
		} else {
			if (!(rowNumber >= 0)) {
				tempobj = $('[name="' + destino + '"]', $formu)[0]; // DESTINO
			} else {
				tempobj = $('[name="' + destino + '"]', $row)[0]; // DESTINO
//			    loger($row);

			}


			// sanitize INTEGERS
			if (tempobj) {
				if (tempobj.getAttribute('type') != undefined)
					var tipo = tempobj.getAttribute('type');
			}

			if ((isNumber(valor) || tipo == 'number' || tipo == 'numeric') && tipo != 'text') {
				var intval = parseInt(valor);
				if (intval - valor == 0) {
					valor = intval;
				} else {
					valor = valor;
		    			try {
						 valor =valor.toFixed(2);
	    				} catch (ex) {
								loger('to fixed exception');
				    	}
					
				}
					//.toFixed(2);
			}

			if (tempobj) {
				if (tempobj.value != valor) {
					tempobj.value = valor;

					if (tempobj.getAttribute('preventLoop') != 'true')
						$(tempobj).change(); // trigger the event manually
				}
				if (rowNumber >= 0) {
					Histrix.calculoTotal(tempobj, null, true);
				}
			}



			xmldatos = formulario.substring(4);

			if (actxml != 'false') {
				setCampo(destino, valor, xmldatos, 'false', tempobj);
			}

			if (formExtra != null ) {

				var idformExtra = formExtra.replace(".", '_');

				tempobj = $('#Form' + idformExtra + ' [name="' + destino + '"]')[0];

				if (tempobj) {
					tempobj.value = valor;

					if (actxml != 'false') {
										
						setCampo(destino, valor, xmldatos, 'false', tempobj);
					}
				}
				// TODO this filtros
				tempobj = $('#FForm' + idformExtra + ' [name="' + destino
						+ '"]')[0];
				if (tempobj) {
					tempobj.value = valor;
				
					if (actxml != 'false') {
						setCampo(destino, valor, xmldatos, 'false', tempobj);
					}
				}

			}
		}
	}
	// calculoForm(formulario);
	if ( actxml != 'false') {
		xmldatos = formulario.substring(4);
		//loger('4481');
		//loger(tempobj);
		setCampo(destino, valor, xmldatos, null, obj);
	}
}

// date functions
// need unification

function straFecha(str) {
	var year = str.substring(0, 4);
	var month = str.substring(4, 6) - 1;
	var day = str.substring(6, 8);

	if (str.substring(2, 3) == '/') {
		day = str.substring(0, 2);
		month = str.substring(3, 5) - 1;
		year = str.substring(6, 11);
	}
	var fecha = new Date(year, month, day);

	return fecha;
}

function aFecha(str) {
	var year = str.substring(6, 11);
	var month = str.substring(3, 5) - 1; // month is 1 less in js
	var day = str.substring(0, 2);
	var fecha = new Date(year, month, day);
	return fecha;
}

function getfecha() {
	var mydate = new Date();
	var year = mydate.getYear();
	if (year < 1000)
		year += 1900;
	var day = mydate.getDay();
	var month = mydate.getMonth() + 1;
	if (month < 10)
		month = "0" + month;
	var daym = mydate.getDate();
	if (daym < 10)
		daym = "0" + daym;
	var fechatxt = daym + "/" + month + "/" + year;
	return fechatxt;
}

function copiavalorcampo(fieldId, origen, xmldatos) {
	var valor = origen.value;
	var idform = xmldatos.replace(".", '_');
	//loger('#Form' + idform);
	var campo = $('[name=' + fieldId + ']', '#IMP' + idform)[0]
			|| $('[name=' + fieldId + ']', '#Form' + idform)[0];
	var destino = campo.name;
	campo.value = valor;
	setCampo(destino, valor, xmldatos, null, campo);
}

// function wrarper 2
function xmlLoader(target, url, options) {
	var opciones = options.suboptions || '';
	var titulo = options.title || '';
	var menuId = options.menuId || '';

	var DefaultLoader = options.loader || 'histrixLoader.php';

	DefaultLoader += '?';

	// Sanitize input
	url = url.replace(/\?/, '');
	return Histrix.loadXML(target, url, opciones, titulo, menuId,
			DefaultLoader, options.reload, options);
}

function cargaInterna(contenedor, contenido, opciones, titulo, xmlprincipal,
		idobj, chkdup) {

	return Histrix.loadInnerXML(contenedor, contenido, opciones, titulo,
			xmlprincipal, idobj, chkdup);
}



// ///////////////////////////
// Funciones de Manejo de Ventanas
// ///////////////////////////

// barra drag igual que la de php, pero en js
function barraDrag(idContenedor, titulo, options) {
	var salida = '';
	if (titulo == undefined)
		titulo = '';
	/*
	 * var minimize = '<button disabled="disabled">'+ '<img
	 * src="../img/icon_minimize_u.png" title="Minimizar" >'+ '</button>';
	 */
	/*
	 * var maximize = '<button class="maximizeButton" >'+ '<img
	 * src="../img/icon_maximize_u.png" title="Maximizar" >'+ '</button>';
	 */
	if (options != undefined && options.maximize == false)
		maximize = '';

	var imgcerrar = '<button class="closeButton" >'
			+ '<img src="../img/icon_close_u.png"  alt="Cerrar" title="Cerrar" >'
			+ '</button>';

	salida += '<div class="barrasup dragBar"  id="dragbar' + idContenedor
			+ '" >';
	salida += titulo;
	salida += '<span class="buttons">' + imgcerrar + '</span>';
	salida += '</div>';

	return salida;
}
// barra resize igual que la de php, pero en js

function barraDragInf(id) {
	var divbar = jQuery('<div ></div>').attr('id', 'dragbar2' + id).addClass(
			'dragBar').css({
		height : '10px',
		margin : '-10px 0'
	});
	return divbar[0];
}

function barrawin(id) {
	var imgresize = '<img src="../img/resize.gif"  alt="tamaño" title="Cambiar Tamaño"/>';
	var divbar = jQuery('<div ></div>').attr('id', 'HLP' + id).addClass(
			'barrainf');
	divbar.html('<span class="resize" id="_resize_' + id
			+ '" onmouseup="Histrix.calculoAlturas(\'' + id + '\', true);" >'
			+ imgresize + '</span>');
	return divbar[0];
}

// //
// Funciones de Validacion y manejo de contenido de valores en el formulario
//

function getDestiny(sourceObject, id,  xml) {

	var field = $('[name='+ id + '][form=Form' + xml + ']' , sourceObject.closest('[instance]') );


	// to find object as inner tables
	if (field.length == 0) {
		field = $('#'+ id  , sourceObject.closest('[instance]')) ;

	}
	/*
	var field = $('#' + formDestino + ' [name="' + id + '"]')[0];

	if (field == undefined) {
		// Inline subForms
		field = $('[name="' + id + '"]', '#TRForm' + idxml)[0];
	}


	if (field == undefined) {
		field = $('[name="' + id + '"], [linkfname="' + id + '"]', '#' + formDestino)[0];
	}
	*/

	if (field.length == 0){
		loger('Error finding field: '+id);
	}



	return field;
}

function actualizarCombo2(obj, idCampo, Destino, xml, form, options) {
//	loger('actualizo combo2: ' + idCampo + ' Dest: ' + Destino);

	var defaultOptions = {
		filter : true
	}
	var formOptions = $.extend(defaultOptions, options);

	// var myregexp = new RegExp('.xml',"g");
	var idxml = xml.replace(".", '_');
	obj = $(obj);
	var parentForm = obj.closest('FORM');
//	var instance = parentForm.attr('instance');
	var instance = obj.closest('[instance]').attr('instance');
	/*
	if (parentForm.attr('xml') != xml && xml != '' || instance == undefined) {
		// get instance from object
		instance = $('[name="' + idCampo + '"]', parentForm).closest('[instance]').attr('instance'); 
		//instance = obj.closest('[instance]').attr('instance');
	}
	*/
	if (obj[0]) {

		if (isHelpOpen(obj[0])) {
			return false;
		}
		// loger(obj);
		var valor = obj.val();
		var idform = obj.attr('form');
		form = obj[0].form;
		var dl = Destino.length;
		var param = {};
		for (i = 0; i < dl; i++) {
			param.destino = Destino[i];
			param.valor = valor;
		}
		param.campo = idCampo;
		var formDestino = 'Form' + idxml;

		if (idform == 'FForm' + idxml) {
			formDestino = idform;
		}

		

		var campoDestino = getDestiny(obj, idCampo, xml);

		/*
		loger(campoDestino);
		loger(instance);
		*/
		var instance = campoDestino.closest('[instance]').attr('instance');

		if (campoDestino && instance != undefined) {

			var objDestino = $(campoDestino.parentNode);
			// $(campoDestino).remove();
			var xmlOrig = $(form).attr('original');
			var xmlorigvar = '';
			if (xmlOrig)
				xmlorigvar = '&xmlOrig=' + xmlOrig;

			// objDestino
			var tempObject = jQuery('<div id="tempid" />');
			param['__winid'] = Histrix.uniqid;

			tempObject.load('refrescaCampo.php?xmldatos=' + xml + xmlorigvar
					+ '&instance=' + instance, param, function () {
				// get old Value
				var $destinationField = $(campoDestino);

				var value = $destinationField.attr('htxValue');

				var inputTemp = tempObject.get(0).firstChild;
				// loger('volvio eso del refresca');
				// loger(inputTemp);
				var hadfocus = false;
				var $inputTemp = $(inputTemp);

				if (value != undefined) {
					$inputTemp.val(value);
					// loger('cuyo valor es:'+value);
					changeSelectValue(inputTemp, value);
				}

				if ($destinationField.is(':focus')) {
					hadfocus = true;
					var currentValue = $destinationField.val();
				}

				$destinationField.replaceWith($inputTemp);

//				$destinationField.css('border', '1px solid red');

				if ($inputTemp.hasClass('refresh') || hadfocus) {
					$inputTemp.focus().val(currentValue);


				}

				$inputTemp.change();

				$('.searchButton' ,$inputTemp.closest('form')).click();
				tempObject.remove();

				// Histrix.registroEventos(idxml);

			});
		}
		/*
		 * if (formOptions.filter == true){ if (parentForm.attr('tipo') ==
		 * 'filter'){ //if ($('#FForm'+idxml)[0]){ // filtracampos('FForm'+idxml ,
		 * xml , xml); //} } }
		 */

	}
	return true;
}

// Pongo el Foco en el Primer Elemento del Formulario
function foco(IDformu, firstId) {
	var formId = IDformu.replace(".", '_');
	$("#" + formId + " .form :input:visible:enabled:first").focus();

}

// cambia el estado seleccionado del Combo
function selectCombo(combo) {

	var lenCombo = combo.options.length;
	for ( var i = 0; i < lenCombo; i++) {
		combo.options[i].removeAttribute('selected');
	}
	combo.options[combo.selectedIndex].setAttribute('selected', 'selected');
}

// /////////////////////////////
// FUNCIONES DE AYUDA
// /////////////////////////////

function searchAyuda(cadena, idinput, xmlayuda, xmlOrig, instance) {
	// var myregexp = new RegExp('.xml',"g");
	var idxmlayuda = xmlayuda.replace(".", '_');
	var input = $('#' + cadena)[0];
	if (input) {
		var hlpayuda = $('#HLP' + idxmlayuda);
		if (hlpayuda[0]) {
			Histrix.loadingMsg(hlpayuda[0].id, 'AYUDA');
		}
		$('#HLP' + idxmlayuda).load(
				"process.php?instance="+instance+"&xmldatos=" + xmlayuda + "&accion=help&divcont=HLP"
						+ idxmlayuda + '&idinput=' + idinput + '&del_filtro='
						+ idinput + '&xmlOrig=' + xmlOrig, {
					cadena : input.value,
					'__winid':Histrix.uniqid
				}, 
				function (){
	    				    Histrix.hideMsg();
					Histrix.calculoAlturas(xmlayuda);
				}
				);
	}
}

function paginar(xmldatos, page, xmlOrig, instance) {
	var idxmldatos = xmldatos.replace(".", '_');
	Histrix.loadingMsg(idxmldatos, ' Page: ' + (page + 1));

	$('#' + idxmldatos).load(
			"refrescaCampo.php?xmlpadre=" + xmlOrig
					+ "&instance=" + instance, {
				pagina : page,
				'__winid':Histrix.uniqid
			});
}


function ObtengoValoresForm(IDformu, encodear, type, inner) {

	// var formuid = $(IDformu).attr('id');
	var vars = "";
	// var vars2 = "";
	var dataObject = {};

	var comma = '';
	var amp = '';

	var inputElements = $(':input', IDformu);

	// var inputElements= $('#'+formuid +' :input')

	var fl = inputElements.length;
	for (i = 0; i < fl; i++) {

		var tempobj = inputElements[i];

		if (tempobj != null) {
			// No saco los valors de los subformularios
			// if (tempobj.hasAttribute('form')) continue;
			// Me fijo si es un contenedor interno
			if (!inner){
				if ($(tempobj).closest('.continternos').length != 0){
					continue;
				}
			}

			if (tempobj.type != "button" && tempobj.type != "reset") {
				var valor = tempobj.value;

				// if (encodear)
				// valor = encodeURIComponent(tempobj.value);

				if (tempobj.type == "checkbox") {
					if (tempobj.checked == true)
						valor = 1;
					else
						valor = 0;
				}

				if (tempobj.type == "radio") {
					if (tempobj.checked) {
						valor = tempobj.value;

						// alert(tempobj.id + '|' + valor);
					} else
						continue;
				}

				// geomap data
				if (tempobj.getAttribute('ctype') != undefined) {
					var strPoints = '';
					var mapId = 'GEOMAP' + $(tempobj).attr('id');

					if (Histrix.polygonControl[mapId]) {

						var storage = Histrix.polygonControl[mapId].storage[0];

						for ( var i = 0; i < storage.geometry.getVertexCount(); i++) {
							var coord = storage.geometry.getVertex(i);
							strPoints += coord.lat() + ',' + coord.lng() + '|';
						}
					}
					if (Histrix.markerControl[mapId]) {
						// Store 1st Point Only
						var storage = Histrix.markerControl[mapId].storage[0];
						var coord = storage.geometry.getPoint();
						strPoints += coord.lat() + ',' + coord.lng() + '|';
					}

					valor = strPoints;
				}

				if (tempobj.name != '') {
					if (type == 'quoted') {
						vars += amp + tempobj.name + '="'
								+ encodeURIComponent(valor) + '"';
					} else
						vars += amp + tempobj.name + "=" + valor;

					if (tempobj.name != '') {

						dataObject[tempobj.name] = encodeURIComponent(valor);
						// vars2 += comma + tempobj.name + ":\"" +
						// encodeURIComponent(valor) + '\" ';
						//
						// comma = ', ';
						// amp = '&';
					}
				}

			}
		}
	}

	if (type == 'obj') {
		return dataObject;
		// eval('vars={'+ vars2 +'}');
	}

	return vars;
}

function ayudaFicha( triggerObject) {
	
	var vars = "";

	var hasValues = false;
	
	var $formu = $(triggerObject).closest('[instance]');
	var instance = $formu.attr('instance');

	var formu 	 = $formu[0];


	if (formu) {
	    
	    instance = $formu.attr('instance');

		$(':input[type=text],:input[type=number]',$formu).each(function(){
			if ($(this).val() != ''){
				hasValues=true;
			}
		});
		

		if (!hasValues){
			var message = 'Ingrese algun valor en un campo para buscar';
			jQuery('<div><p><span class="ui-icon ui-icon-alert" style="float: left; margin: 0 7px 20px 0;"></span>'+message+'</p></div>').dialog(
			    {modal:true,
			    resizable: false,
			    buttons: {
		                Ok: function() {
	            		        $( this ).dialog( "close" );
			            }
		                }
			    }

			);
			return false;
		}

		Histrix.showMsg('Buscando...');
		loger(formu);
		vars = ObtengoValoresForm(formu, true, 'obj');

		vars['__winid']=Histrix.uniqid;
		$.post('getXmlData.php?ayudaFicha=true&instance=' + instance, vars,
			function(xmldata){
				llenoCamposdesdeXML(xmldata);
			});

	} 
	/*else {
		var div = $('TR' + formu.id)[0];
//		vars = recorroNodo(div, 'obj');
		vars = Histrix.getNodeData(div);
	}
	*/

}

function popAyuda(xmlprincipal, idinput, xmlayuda, e, postOptions) {

	
	var defaultOptions = {
		xmlform : '',
		helpStyle : '',
		xmlOrig : '',
		instance : ''
	}
	var options = $.extend(defaultOptions, postOptions);
         
	var xmlform = options.xmlform;
	var helpStyle = options.helpStyle;
	var xmlOrig = options.xmlOrig;

	var key = e.keyCode;

	if (key == 113 || e.type == 'click') {

		var inp = $(idinput);

		var dato = inp.val();

		var pos = inp.position();

		if (pos) {
			var tempX = pos.left;
			var tempY = pos.top;
		}

		// var myregexp = new RegExp('.xml',"g");
		var idxmlayuda = xmlayuda.replace(".", '_');
		var idxmlform = xmlform.replace(".", '_');
		var idxmlprincipal = xmlprincipal.replace(".", '_');
		var idxmlOrig      = xmlOrig.replace(".", '_');

		
		
		if ($('#HLP' + idxmlayuda)[0] == undefined) {

			var supra = $('#Det' + idxmlform);

			if ($('#' + idxmlprincipal)[0]) {
				supra = $('#' + idxmlprincipal);
			}
			if (supra[0] == undefined)
				supra = $('#Det' + idxmlprincipal);
			if (supra[0] == undefined)
				supra = $('#' + idxmlform);
			if (supra[0] == undefined)
				supra = $('#Form' + idxmlform);

			// var idxmlOrig = idxmlprincipal.substring(3);
			// var xmlOrig = xmlprincipal.substring(3);

			if (supra[0] == undefined)
				supra = $('#Show' + idxmlOrig);


			if (supra[0] == undefined)
				supra = Histrix.Supracontenido;

			var newdiv = jQuery('<div origen="' + inp.attr('id') + '"></div>')
					.attr('id', 'HLP' + idxmlayuda).addClass('ayudaint');
			// .draggable({handle:'#dragbarHLPHLP'+idxmlayuda});

			if (helpStyle != '') {
				newdiv.attr("style", helpStyle);
			}
			if (supra.attr('original') != undefined) {
				var idorig = supra.attr('original');
				idorig = idorig.replace(".", '_');

				var orig = $('#Show' + idorig);
				if (orig[0])
					supra = orig;
			}

			if (inp[0]) {
				xmlform = inp.attr('form');

				supra.append(newdiv);

				var h = document.body.offsetHeight;
				var w = document.body.offsetWidth;

				if (tempY + newdiv.offsetHeight > h)
					tempY -= (newdiv.offsetHeight + (inp.offsetHeight * 3));
				if (tempX + newdiv.offsetWidth > w)
					tempX = w - newdiv.offsetWidth - 10;

				newdiv.css({
					top : (tempY) + 'px',
					left : (tempX + inp.width()) + 'px'
				});

				// TODO: REMOVE ID REFERENCE

				if (dato != '') {
					Histrix.loadingMsg(newdiv[0].id, 'AYUDA');
					$(newdiv).load(
							"process.php?xmldatos=" + xmlayuda
									+ "&accion=help&divcont=HLP" + xmlayuda
									+ '&idinput=' + inp[0].name
									+ '&del_filtro=' + inp[0].name
									+ '&xmlOrig=' + xmlOrig + '&instance='
									+ options.instance, {
								cadena : dato,
								'__winid':Histrix.uniqid
							}, 
							function (){
							    Histrix.hideMsg();
								Histrix.calculoAlturas(xmlayuda);
							}							
							);

				} else {
					var str = 'str' + idxmlayuda;
					var search = 'searchAyuda(\'' + str + '\', \''
							+ inp[0].name + '\', \'' + xmlayuda + '\' , \''
							+ xmlOrig + '\',\''+ options.instance +'\');';

					var html = '<div class="popayuda" ><input id="'
							+ str
							+ '" type="search" onChange="'
							+ search
							+ '"  ><button type="button" value="buscar" name="buscar" onClick="'
							+ search
							+ '">Buscar</button><button type="button" '
							+ 'value="cerrar" name="cerrar" '
							+ 'onClick="cerrarVent(\'' + newdiv[0].id
							+ '\');">Cerrar</button></div>';
					newdiv.html(html);
					$('#' + str).focus();

				}

			}
		}
	}
	return false;
}

function addInnerGraphWindow(xmlprincipal, title, innerclass) {

	var contenedor = xmlprincipal.replace(".", '_');
	// var contenedor = xmlprincipal;
	var supra = $('#DIV' + contenedor);
	if (!supra[0]) {
		supra = Histrix.Supracontenido;
	}

	var newdiv = jQuery('<div></div>').attr("id", 'PRN' + contenedor).addClass(
			innerclass);
	var internos = $('.' + innerclass);
	if (internos[0]) {
		var offset = internos.length * 15;
	}
	supra.append(newdiv);
	newdiv.css({
		top : 20 + offset + 'px',
		left : 20 + offset + 'px'
	});

	var barra = barraDrag('PRN' + contenedor, title);
	newdiv.html(barra);

	var contewin = jQuery('<div></div>');
	var now = new Date();
	contewin.html(
			'<img alt="' + title + '" src="tree_graph.php?xml=' + xmlprincipal
					+ '&' + now.getTime() + '" />').attr("id",
			'DIV' + contenedor).addClass('contewin');
	newdiv.draggable({
		handle : '#dragbarPRN' + contenedor
	}).append(contewin);
}

uniqid = (function () {
	var id = 0;
	return function () {
		return id++;
	};
})();

function changeSelectValue(obj, val) {
	// tempobj.options[obj.selectedIndex]
	// Refresh object

	obj = $(obj)[0];
	var arrayopciones = obj.options;
	if (arrayopciones)
		for ( var index = 0; index < arrayopciones.length; ++index) {
			var item = arrayopciones[index];

			if (item.value == val)
				item.setAttribute('selected', 'selected');
			else
				item.removeAttribute('selected');
		}
}

// Cargo el valor de la primera columna de la tabla en el campo destino
function cargoValor(Fila, idinput, xmlOrigen) {
	// loger('cargoValor');
	var field = null;
	var idxmlOrigen = xmlOrigen.replace(".", '_');

	// Event.stopObserving(document, 'keypress');

	var cols = Fila.getElementsByTagName("td");
	var hlpxml = $(Fila.parentNode.parentNode).attr('xml');
	var divayuda = 'HLP' + hlpxml;
	//loger('divayuda' + divayuda);

	var newValue = $(cols[0]).text();

	var orig = undefined;
	if ($('#Form' + idxmlOrigen)[0]) {
		form = $('#Form' + idxmlOrigen);
		field = $('#Form' + idxmlOrigen + ' [name="' + idinput + '"]');
		field.val(newValue);

	}
	if ($('#FForm' + idxmlOrigen)[0]) {
		form = $('#FForm' + idxmlOrigen);
		field = $('#FForm' + idxmlOrigen + ' [name="' + idinput + '"]');
		field.val(newValue);
		orig = form.attr('original');
	}

	if ($('#TRForm' + idxmlOrigen)[0]) {
		form = $('#TRForm' + idxmlOrigen);
		field = $('#TRForm' + idxmlOrigen + ' [name="' + idinput + '"]');
		field.val(newValue);

	}

	if ($('#' + idinput)[0]) {
		field = $('#' + idinput);
		var form = $(field[0].form);
	}

	if ($('#' + divayuda).attr('origen')) {
		var idField = $('#' + divayuda).attr('origen');
		field = $('#' + idField);
		var form = $(field[0].form);

	}

	if (field != null)
		field.val(newValue).focus().blur();


/*
	if (form != undefined) {
		orig = form.attr('original');

		if (orig != form.attr('name') && orig != undefined) {
		llenoCabecera(form.attr('id'), xmlOrigen, orig);
			
		}
	}

*/

	// field[0].value=newValue;

	field.attr('valid', true);

	cerrarVent(divayuda);

	if (field.attr('onchange')) {
		Histrix.buscando = false;
		field.change();
	}

	
	
}

String.prototype.unescapeHtml = function () {
	var temp = document.createElement("div");
	temp.innerHTML = this;
	if (temp.childNodes[0]) {
		var result = temp.childNodes[0].nodeValue;
		temp.removeChild(temp.firstChild)
		return result;
	}
	return this;
}

/* wrarper function */

function fillForm(Fila, xmlupdate) {

	var tbody = $(Fila.parentNode.parentNode);
	// var idForm = tbody.attr('idForm');
	var xmldatos = tbody.attr('xml');
	var type = tbody.attr('type');
	var idForm = 'Form' + xmldatos;
	llenoForm(Fila, idForm, xmldatos, xmlupdate, type);
}

/* lleno el ABM con los datos del registro Seleccionado de la tabla */
/* Select Data from Table Row and put it into the FROM */
function llenoForm(Fila, idForm, xmldatos, xmlupdate, tipo, dataObject) {

	 //loger('LlenoForm: '+idForm+' '+ xmldatos+' '+ xmlupdate+' '+ tipo);

	// var myregexp = new RegExp('.xml',"g");
	var idxmldatos = xmldatos.replace(".", '_');
	var idForm2 = idForm.replace(".", '_');
	$divform = $('#DIVFORM' + idxmldatos);
	
	if (!$divform.is(':visible')) {
		$divform.slideDown();
	}


	// var cols = Fila.getElementsByTagName("td"); // Obtengo las Columnas

	var cols = $("> td", Fila); // get ells without inner tables

	var Formu = $("#" + idForm2+':visible')[0]; // Get Form

	var hayformu = false;
	if (Formu)
		hayformu = true;

	var div = $("#TR" + idForm2)[0];
	// var offset = 0;
	var arefrescar = [];
	var l = 0;
	var content;
	var columnas = cols.length;

	if (dataObject != undefined) {
		cols = dataObject;
		columnas = Object.keys(cols).length;
	}
	var tempobj;
	var obj;
	var boton;
	var botonborrar;
	var datos = '';
	var val;
	var fieldValues = {};
	var parentTable = getParent(Fila, 'TABLE');

	// replace instance _aux_ for fill form
	
	if ($(parentTable).attr('instance'))
		var instance = $(parentTable).attr('instance').replace('_aux_', '');


	var colgroupCols = $('col', parentTable);
	var destino = '';

	if (hayformu) {

		var valor;
		var i = 0;

		for (i = 0; i < columnas; i++) {
			var celda = cols[i];

			if (dataObject != undefined) {
				instance = dataObject.instance;
				var count = 0;
				for ( var key in cols) {

					if (count == i && cols.hasOwnProperty(key)) {
						destino = key;
						valor = cols[key];

					}
					count++;
				}
			}

			// Check if there is an Inner Table
			if (celda != undefined)
				if (celda.firstChild && celda.firstChild.localName == 'TABLE') {
					var innerTable = cols[i].firstChild;
					if (innerTable.className == 'autofields') {
						var tbody = innerTable.getElementsByTagName("tbody");

						var Filas = tbody[0].getElementsByTagName("tr");
						var Nfilas = Filas.length;
						for ( var n = 0; n < Nfilas; n++) {
							var fila = Filas[n];
							datos += Histrix.getRowValues(fila, Formu);
						}
					}
				}

			// Normal load of Form
			if (celda != undefined) {
				if ($(celda).attr('valor')) {
					valor = $(celda).attr('valor');
				} else {

					valor = $(celda).html().unescapeHtml();
					if (valor == null)
						valor = celda.innerText;
				}

				if (cols[i].getAttribute('valauto') != undefined) {
					valor = cols[i].getAttribute('valauto');
				}

				destino = '';
				if (colgroupCols[i])
					destino = colgroupCols[i].getAttribute('campo')
				else
					destino = cols[i].getAttribute('campo');

			}
			// var fl = Formu.length;

			// Prototype Select Element method
			var tempobjArray = $('[name="' + destino + '"]', Formu);

			// add Values to Object to be send after
			fieldValues[destino] = encodeURIComponent(valor);

			tempobj = tempobjArray[0];
			$tempobj = $(tempobj);
			$tempobj.attr('htxValue', valor);

			if (tempobj) {
				var oldValue = tempobj.value;

				if (tempobj.type == "radio") {
					for ( var ij = 0; ij < tempobjArray.length; ij++) {
						if (valor == tempobjArray[ij].value)
							tempobjArray[ij].checked = true;
						else
							tempobjArray[ij].checked = false;
					}
				} else
					tempobj.value = valor;

				if (tempobj.tagName == "SELECT") {
					// loger('en el caso de selects.') ;
					if (oldValue != tempobj.value)
						tempobjArray.change();

					changeSelectValue(tempobj, valor);
					// loger('fin en el caso de selects.') ;
				}

				// recurrencerule

				if ($tempobj.hasClass('recurringrule')) {
					if (valor != '')
						$tempobj.recurrenceinput().loadData(valor);
				}

				if (tempobj.getAttribute('escolor') != undefined) {
					tempobj.style.backgroundColor = valor;
				}
				if ($tempobj.attr('image') == 1) {
					$tempobj.change();
				}
				if (tempobj.getAttribute('ctype') != undefined) {
                                                           /*
					var mapId = 'GEOMAP' + $(tempobj).attr('id');

					if (Histrix.polygonControl[mapId]) {
						Histrix.geoDrawPolygon(mapId, valor);
					}
					if (Histrix.markerControl[mapId]) {
						Histrix.geoDrawPoint(mapId, valor);
					}
				
					*/
					var point = valor.split(',');
					var myLatlng = new google.maps.LatLng(point[0],point[1]);


					var marker = new google.maps.Marker({
					      position: myLatlng,
					      map: Histrix.mymap,
					      title: "TEST"
					});

					Histrix.mymap.panTo(myLatlng);
    			        	//marker.setMap(map);  
					
				}

				if (tempobj.type == "checkbox") {
					if (valor == 1)
						tempobj.checked = true;
					else
						tempobj.checked = false;

					if ($tempobj.hasClass('activate')){
						$tempobj.change();
					}
				}
		//				$tempobj.change();

				if (tempobj.getAttribute('internal_class')) {

					switch (tempobj.getAttribute('internal_class')) {
					case 'simpleditor':
						$('#' + tempobj.id).wysiwyg('setContent',
								cols[i].innerHTML);
						break;
					}
				}
				if (tempobj.getAttribute('pintado') != undefined) {
					tempobj.select();
				}

				datos += destino + '=' + valor + '&';

			}

			if (cols[i])
				if (cols[i].getAttribute('obj') != undefined) {
					var xmlrefrescar = cols[i].getAttribute('obj').toString();
					// almaceno los contenedores a llenar
					if (xmlrefrescar != '') {
						arefrescar[l] = xmlrefrescar;
						l++;
						datos += destino + '=&';
					}
				}
		}
		boton = $('[name=Grabar]', Formu)[0];
		botonborrar = $('[name=delete]', Formu)[0];

	} else {
		// Cuando es un subformulario y por lo tanto no tengo FORM
		// recorro las columnas

		val = null;
		// offset = 0;
		var k = 0;

		for (k = 0; k < columnas; k++) {

			var fieldName = '';
			if (colgroupCols[k])
				fieldName = colgroupCols[k].getAttribute('campo')
			else
				fieldName = cols[k].getAttribute('campo');

			var objinput = $('#' + idxmldatos + '[instance="'+instance+'"] [name="' + fieldName + '"]');

			// fallback if not found
			if (objinput.length == 0){
				objinput = $('#' + idxmldatos + ' [name="' + fieldName + '"]');
			}

			// caMBIO LEYENDA Boton Crear
			objinput.each(function (index, elem) {
				val = elem;
				var cell = $(cols[k]);
				
				if (val) {
					if (cell.attr('valor') != undefined) {
						valor = cell.attr('valor');
					} else {
						valor = cell.html(); // .innerHTML.unescapeHTML();

					}
					if (val.type != "text" || valor != '') {

						val.value = valor;

					} else {
						if (document.all) {
							val.value = cols[k].innerText;
						} else {
							val.value = cols[k].textContent;
						}
					}
					if (val.getAttribute('pintado') != undefined) {
						$(val).select();
					}
					// vars+='&'+val.name+'='+val.value;
				}
			});
		}
		// caMBIO LEYENDA Boton Crear

		obj = $('#' + idxmldatos + ' [type="button"]');
		boton = null;
		botonborrar = null;
		obj.each(function (index, elem) {
					if (elem.name == 'Grabar') {
						boton = elem;
						boton.innerHTML = '<img src="../img/filesave.png" alt="'
								+ Histrix.i18n['save']
								+ '"  />'
								+ Histrix.i18n['save'];
						boton.setAttribute('accion', 'update');
					}
					if (elem.name == 'delete') {
						botonborrar = elem;
						botonborrar.disabled = false;
					}
				});
	}
	//loger('CHANGE');
	if (Fila != null)
		Histrix.selectRow(Fila);


	/* Los Valores los almaceno en el objeto contenedor de datos */

	obj = $('#' + idForm2 + ' [type="button"]');

	boton = null;
	botonborrar = null;
	obj.each(function (num) {
		var elem = obj[num];
		if (elem.name == 'Grabar') {
			boton = elem;
			boton.innerHTML = '<img src="../img/filesave.png" alt="'
					+ Histrix.i18n['save'] + '"  />' + Histrix.i18n['save'];
			boton.setAttribute('accion', 'update');
		}

		if (elem.name == 'delete') {
			botonborrar = elem;
			botonborrar.disabled = false;
		}
	});
	Histrix.hideMsg();
	// var vars = "";
	if (hayformu) {
		vars = fieldValues;
		// vars = ObtengoValoresForm(Formu, true, 'obj');
		
	} else {
//		vars = recorroNodo(div, 'obj');
		vars = Histrix.getNodeData(div);
	}

	// loger(fieldValues);
	// loger(vars);
	if (xmlupdate != 'false') {
		vars['__winid']=Histrix.uniqid;

		$.post("setData.php?_show=false&modificar=true&instance=" + instance, vars, function () {
			refrescarInternos(arefrescar, xmldatos, vars);
		});

	}

	if (boton) {
		boton.innerHTML = '<img src="../img/filesave.png" alt="'
				+ Histrix.i18n['save'] + '"  />' + Histrix.i18n['save'];
		boton.setAttribute('accion', 'update');

		if (botonborrar)
			botonborrar.disabled = false;
	}
}
/*
function refrescaCampo(xml, campo2) {
	var $field = $('#' + campo2);
	var dest = $field.parent()[0];
	var form = $field[0].form;
	var xmlOrig = $(form).attr('original');
	var xmlorigvar = '';
	if (xmlOrig)
		xmlorigvar = '&xmlOrig=' + xmlOrig;
	$field.remove();
	$(dest).load("refrescaCampo.php?xmldatos=" + xml + xmlorigvar, {
		campo : campo2,
		'__winid':Histrix.uniqid
	});
}
*/
function refrescarInternos(arefresh, xmlpadre, vars) {
	// var myregexp = new RegExp('.xml',"g");
	var lref = arefresh.length;
	if (lref > 0) {
		var i = 0;
		for (i = 0; i < lref; i++) {
			var str = arefresh[i].replace(".", '_');
			var destino = $('#' + str)[0];
			if (destino) {
			//	vars['__winid']=Histrix.uniqid;
				var innerInstance = $('#' + str + ' [instance]').attr('instance');
				$('#' + str).load(
						"refrescaCampo.php?xmlpadre=" + xmlpadre 
						+ "&instance=" + innerInstance 
						+ "&xmldatos=" + arefresh[i], vars);
			}
		}
	}
}

function llenoCabecera(idForm, xmldatossub, xmldatos, field) {
	var Formu = $('#' + idForm)[0];
	// Los Valores los almaceno en el objeto contenedor de datos
	
	if (field != undefined){
		Formu = $(field).closest('form');

		xmldatos = Formu.attr('xmlorig');
	}


	if ($(Formu).attr('tipo') != 'cabecera')
		return false;

	var vars = ObtengoValoresForm(Formu, false, 'obj');
	//alert('lleno Cabecera '+idForm+ ' '+xmldatossub + ' ' +xmldatos );
	

	//alert('valores');
	var instance = $(Formu).attr('maininstance');
	//vars['__winid']=Histrix.uniqid;
	$.post("setData.php?_show=false&xmldatossub=" + xmldatossub + '&instance=' + instance, vars);

}

function confirmacion(mensaje) {
	return confirm(mensaje);
}

function ObjetoInput(NomCampo, operador, Val) {
	this.NombreCampo = NomCampo;
	this.Operador = operador;
	this.Valor = Val;
}

function onoffAtt(attrib, attrcon, attrval, form) {
	$('#' + form + ' [' + attrcon + '=' + attrval + ']').each(function () {
		var $this = $(this);
		if ($this.attr(attrib) == 'true')
			$this.attr(attrib, 'false');
		else
			$this.attr(attrib, 'true');
	});
}

function addfiltro(xmldatos, xmlOrig) {
	var idxmldatos = xmldatos.replace(".", '_');

	var autocampo = $('#_Autofiltro' + idxmldatos).val();
	var autooperador = $('#_AutoOperador' + idxmldatos).val();
	// Saco los filtros previos
	cerrarVent('Filtros' + idxmldatos);
	var filtros = '#IMP' + idxmldatos;
	if ($('#Filtros' + idxmldatos)[0])
		filtros = '#Filtros' + idxmldatos;
	else {
		// create Node
		filtros = jQuery('<div></div>').attr('id', 'Filtros' + idxmldatos)
				.addClass('filtro');
		$('#IMP' + idxmldatos).prepend(filtros);
	}
	var instance = $('#Filtros' + idxmldatos).closest('[instance]').attr('[instance]');
	$(filtros).load(
			"process.php?addfiltro=addfiltro&xmldatos=" + xmldatos + '&instance=' + instance
					+ '&xmlOrig=' + xmlOrig + '&autocampo=' + autocampo
					+ '&autooperador=' + autooperador, {'__winid':Histrix.uniqid});
}

function delautofiltro(xmldatos, truid, uid, xmlOrig) {
	// var myregexp = new RegExp('.xml',"g");
	var idxmldatos = xmldatos.replace(".", '_');
	// Saco los filtros previos
	cerrarVent(truid);
	var instance =$(idxmldatos).closest('[instance]').attr('[instance]');
	$(idxmldatos).load(
			"process.php?delfiltros=delfiltros&removerfiltro=removerfiltro&xmldatos="
					+ xmldatos + '&uidfiltro=' + uid + '&xmlOrig=' + xmlOrig + '&instance=' + instance, {'__winid':Histrix.uniqid});
}

function autoComplete(event, data, idForm, rowNumber) {
	//loger('autocomplete jQuery');
	var localForm = event.target.form;
	var node = {};

	// first Pass JUST Set Values
	for ( var i = 1; i < data.length; i++) {
		// loger(data);
		eval(data[i]);
		if (node.name != undefined) {
			if (rowNumber != null) {
				var $row = $(event.target).closest('tr');
				$('[name="' + node.name + '"]', $row).val(node.value);
			} else {
				$('[name="' + node.name + '"]', localForm).val(node.value);
			}
		}
	}

	// seccond Pass trigger change method
	for (i = 1; i < data.length; i++) {
		// loger(data);
		eval(data[i]);
		if (node.name != undefined) {
			if (rowNumber != null) {
				$row = $(event.target).closest('tr');
				$('[name="' + node.name + '"]', $row).change();
			} else {
				$('[name="' + node.name + '"]', localForm).change();
			}
		}
	}

}

function toggleID(id) {
	$('#' + id).toggle();
}

function filtracampos(IDformu, tabla, xmldatos, xmlOrig, instanceinput) {

	// var myregexp = new RegExp('.xml',"g");
	var filterDiv = 'Filtros' + tabla.replace(".", '_');
	if (xmldatos == undefined)
		xmldatos = tabla;
	var xml = xmldatos.replace(".", '_');

	Histrix.showMsg('filtrando');
	var arrayCampos = []; // 20 de valor?

	var filterForm = $('#' + IDformu);

	if (!filterForm[0]) {
		loger('no hay filtros');
		filterForm = $('#' + filterDiv);
	}

	if (!filterForm[0]) {
		filterForm = $('#F' + filterDiv);
	}

	var instance = filterForm.attr('instance');

	if (instance == undefined)
		instance = instanceinput;

	if (xmlOrig == undefined) {

		xmlOrig = filterForm.closest('FORM').attr('xml');
	}

	var i = 0;
	var error = false;
	$(':input', filterForm).each(
			function () {
				var tempobj = $(this);

				var oblig = $(tempobj).attr('oblig');

	          if (tempobj.val() == '' && oblig == 'true'){
                 //     loger($(tempobj));

  		         Histrix.alerta('obligatorio', tempobj);
  		         
                 error = true;
              }
                                   

				if (tempobj[0].type != "button" && tempobj[0].type != "reset"
						&& (tempobj.val() != '' || oblig != '')) {

					if (tempobj.attr('name') != "operador"
							&& tempobj.attr('deshab') != 'true'
							&& tempobj.attr('name') != '') {
						/* TEST !!!!!!!!!! donde hay radios? */

						var campo = tempobj.attr('name');
						var valor = tempobj.val();

						if (tempobj[0].type == "radio") {

							if (tempobj[0].checked != true)
								return;

						}

						if (tempobj[0].type == "checkbox") {
							if (tempobj[0].checked)
								valor = '1';
							else
								valor = '0';
						}

						if (valor != '' || oblig != '') {
							var operador = '=';
							if (tempobj.attr('operador') != undefined) {
								operador = tempobj.attr('operador');
							}

							if (campo != undefined) {
								var arrayObj = new Array(3);

								arrayObj[0] = campo;
								arrayObj[1] = operador;
								arrayObj[2] = valor;
								arrayCampos[i] = arrayObj;

								i++;
							}

						}
					}
				}
			});
    
    if (error){ 
    	Histrix.hideMsg();
  		return;
	}
    
    			
	var miJSON = $.toJSON(arrayCampos);
	tabla = tabla.replace(".", '_');
	Histrix.loadingMsg(tabla, 'Buscando');

	$('#' + tabla).load(
			"setFiltros.php?instance=" + instance, 
			{ mijson : miJSON,
			  '__winid':Histrix.uniqid
			}, function (responseText, textStatus) {
				if (textStatus == 'error'){
					Histrix.loadingMsg(tabla, 'Error de búsqueda' + responseText, false);
				}
				Histrix.hideMsg();
				Histrix.registroEventos(xml);
			});

	// resizeTabla(tabla , xmldatos);
}

function delfiltros(tabla, xmldatos) {
	// var myregexp = new RegExp('.xml',"g");
	var idxmldatos = xmldatos.replace(".", '_');

	$('#FForm' + idxmldatos)[0].reset();
	// Remove unnecesary RoundTrip to server
	// $(tabla).load("process.php?delfiltros=delfiltros&xmldatos="+xmldatos);
}

/* Seteo varios valores de campos del mismo xml */
function setCampos(arraycampos, valores, xmldatos, show, xmlOrig, instance) {
	var lc = arraycampos.length;
	var vars = '';
	for ( var i = 0; i < lc; i++) {
		vars += arraycampos[i] + '=' + valores[i];
	}
	var xmlorigvar = '';
	if (xmlOrig)
		xmlorigvar = '&xmlOrig=' + xmlOrig;
	
	vars['__winid']=Histrix.uniqid;
	/*
	if (instance == undefined){
		alert(instance);
		loger('error' + xmldatos);
		loger(vars);
	}*/

	$.post("setData.php?instance=" + instance + xmlorigvar + "&_show=" + show, vars);
}

/* Seteo los campos en los Arboles */
function setCampo(campo, valor, xmldatos, show, tempfield) {
	var fields = new Array(1);
	var values = new Array(1);
	fields[0] = campo;
	values[0] = valor;

	var xmlOrig = '';
	var instance = undefined;

	if (tempfield) {
		xmlOrig = $(tempfield.form).attr('xmlOrig');

		if (xmlOrig == undefined)
			xmlOrig = $(tempfield.form).attr('original');

		instance = $(tempfield).closest('form').attr('instance');

	}
	else {
		//alert('no tengo objeto');
	}
	
	
	setCampos(fields, values, xmldatos, show, xmlOrig, instance);

}

function setValorCampo(formu, campo, valor, xmldatos) {
	// loger('setValorCampo jQuery');

	$('[name="' + campo + '"]', formu).each(function (index, tempobj) {
		if (tempobj.type == "radio") {
			if (valor == tempobj.value)
				tempobj.checked = true;
			else
				tempobj.checked = false;
		} else
			tempobj.value = valor;
		if (tempobj.type == "checkbox") {
			if (valor == 1)
				tempobj.checked = true;
			else
				tempobj.checked = false;
		}
	})
}

// Actualizo el objeto contenedor con los datos del select.option
function setCampoSilent(select, campo, xmlcabecera, xmldatos) {
	// loger(' call setCampoSilent');

	option = $('option:selected', select)
//.options[select.selectedIndex];
	var valor = $(option).attr(campo);
	var vars = campo + "=" + valor;

	if (xmldatos == '')
		xmldatos = xmlcabecera;

	var $form = $(select).closest('form');
//	var xmlOrig = $form.attr('xmlOrig');
	var instance = $form.attr('instance');

	if (xmldatos != '' && instance != undefined) {
//	vars['__winid']=Histrix.uniqid;		
		$.post("setData.php?_show=false&instance=" + instance + "&xmldatossub=" + xmlcabecera , vars);
	}

	if (instance == undefined){
		loger('FIX: instance undefined line 6036 ' + vars);
	}

	if (!select.getAttribute('noupdate') != undefined) {
		var currentForm = $form;

		var $row = $(select).closest('[o], form').first();

		$('[name="' + campo + '"]', $row).each(function () {
			var $this = $(this);
			var oldValue = $this.val();
			$this.val(valor);
			var tempobj = $this.get(0);
			if (tempobj.type == "checkbox") {
				if (valor == 1)
					tempobj.checked = true;
				else
					tempobj.checked = false;
			}

			if (oldValue != valor) {
				$this.change();
			}

		});

	}

	return true;

}

function setCampoTabla(fila, campo, input, xmldatos, refresco, xmlOrig,
		instance) {
	var valor = undefined;
	if ($(input).is(':checkbox')){
	    valor = input.checked;
	} else {
	    valor = $(input).val();
	}
	var tr = getTR(input);
	var valget = '&getFila=true&fila=' + fila + '&campo=' + campo + '&valor='
			+ valor + '&xmlOrig=' + xmlOrig + '&instance=' + instance;
	if (refresco == false) {
		$.post("setData.php?_show=false"  + valget, {'__winid':Histrix.uniqid});
	} else {
		$(tr).load("process.php?xmldatos=" + xmldatos + valget ,
			{'__winid':Histrix.uniqid}
			,
			// after update row
			
			function (){
				
	                       // recalc table sum  
            		    $('col', $(tr).closest('TABLE')).each(
            		        function (index, element){
	        		    		if ($(element).hasClass('celsum')){
									var $celsum = $(tr.cells[index]);
                            		Histrix.calculoTotal(undefined, undefined, undefined, {cell:$celsum});
	        		    		}
                        	}
                        );
                         
                	}

	
		);
	}
}

function buscar(contenedor, campo, valor, operador, xmldatos, xmlOrig, field) {
	if (valor != '') {
		var instance = $(field).closest('[instance]').attr('instance');
		var vars    = {'__winid':Histrix.uniqid , '__searchField':campo};
		vars[campo] = valor;

		$.post('getXmlData.php?xmldatos=' + xmldatos + '&xmlOrig' + xmlOrig+'&instance='+instance,
				vars, llenoCamposdesdeXML, "xml");
	}
}

function cerrarVent(nombre, element) {


	nombre = nombre.replace(".", '_');

	if (element != undefined){
		var $win 	 = $(element).closest("#" + nombre);
		var instance = $('[instance]:first', $win).attr('instance');
		// because a bug in IE9 must empty element before remove it
		$win.empty().remove();

	} else {
		var $win = $("#" + nombre);
		var instance = $('[instance]:first', $win).attr('instance');
		$win.remove();
	}

	$("#MODAL" + nombre).remove();

 	/*
 	
 	 this breaks help windows.. fix!
	if (instance)
		$.post("delvars.php?instance=" + instance , {'__winid':Histrix.uniqid});
	*/


}

function cerrarVentclase(nombre, excluir) {
	// $("." + nombre).not('#'+excluir).remove();
	// loger('cerrarVentclase jQUERY ');

	var j = document.getElementsByClassName(nombre).length;
	if (j > 0) {
		var arrayDiv = document.getElementsByClassName(nombre);
		for (i = 0; i < j; i++) {
			var aDiv = arrayDiv[i];
			if (aDiv.id != excluir) {
				// Get it's parent
				var pTag = aDiv.parentNode;
				// Remove
				pTag.removeChild(aDiv);
				if (pTag.id == 'PRN' + nombre)
					cerrarVent(pTag.id);
			}
		}
	}
}

function isHelpOpen(input) {
	var isHelp = $('[origen="' + input.id + '"]')[0];
	if (isHelp == undefined) {
		return false;
	} else {
		return true;
	}
}
/*
 * Call function on Help fields
 */
// TODO rename to searchRecord
function buscaregistro(id, nombreCampo, obj, xmldatos, novalidate, e,
		postOptions) {

	// Default windows Options
	var defaultOptions = {
		xmlOrig : '',
		instance : ''
	}
	var options = $.extend(defaultOptions, postOptions);

	
	// Check if there is not a previous open help
	/*
	if (isHelpOpen(obj)) {
		alert('return');
		return false;
	}
	 */
	 /*
	var $form = $(obj).closest('FORM');
	var formName = $form.attr('name');

	var formContainer = '';
	
	if (formName != 'Form' + xmldatos) {
		formContainer = '&formContainer=' + formName;
	}
	*/
	var valor = $(obj).val();
	// set Invalid before search
	$(obj).attr('valid', 'false');

	if (valor != '') {
		Histrix.showMsg('Buscando...');
		
		var vars = {'__help':nombreCampo};
		vars[nombreCampo] = valor;
		if (Histrix.buscando == false) {

			$.post('getXmlData.php?instance=' + options.instance 
					//+ formContainer 
					, vars, 
					function(xmldata){
						llenoCamposdesdeXML(xmldata, obj);
					}, 'xml');

		}
		Histrix.buscando = true;
	} else {
		$(obj).attr('valid', 'true');
	}
}

function llenoCamposdesdeXML(xml, source) {

	// loger('read xml');

	var vars = "";
	var xmldatos;
	var xmlfile;
	var formname;
	var tempobj;
	var xmlinstance;
	Histrix.hideMsg();
	var obtengo = true;
	var autosube = false;
	var form = {};
	var hayform;

	var $resultado = $(xml).find('resultado');
	var instance = $resultado.attr('instance');
	var parentInstance = $resultado.attr('parentInstance');
	var resultado = $resultado[0];

	if (resultado != null) {
		if ($resultado.attr('vacio') !== undefined) {
			// var vacio = resultado.getAttribute( 'vacio' ).toString();
			// if (vacio == 'true') Histrix.alerta('NO SE ENCONTRARON
			// COINCIDENCIAS');
			obtengo = false;
		}
		var ayudaficha = $resultado.attr('ayudaficha');
		if (ayudaficha !== undefined) {

			var origen = $resultado.attr('xmlorigen');
			if (ayudaficha == 'true')
				Histrix.alerta('NO SE ENCONTRARON COINCIDENCIAS', origen);
			obtengo = false;
		}

		// get FORM
		hayform = false;
		if (parentInstance != ''){
			form = $('form[instance="'+parentInstance+'"]');//, DIV[instance='+parentInstance+']');
			if (form.length == 0){
			    form = $('DIV[instance="'+parentInstance+'"]');
			}
			instance = parentInstance;

		} else {
			form = $('form[instance="'+instance+'"], DIV[instance="'+instance+'"]');
		}

		if (form.length != 0) {
			hayform = true;
		}			
		/*
		// Modificacion para que funcione las ayudas con los filtros
		// No me traia el detalle sino
		if ($('#FForm' + xmldatos)[0]) {
			hayform = true;
			form = $('#FForm' + xmldatos);
		}

		*/


		var error = $resultado.attr('error');
		if (error !== undefined) {
			xmldatos = $resultado.attr('xmlorigen');


			// Histrix.alerta('VALOR INCORRECTO', $(error));
			// setTimeout("$('"+error+"').focus();",1);
			//
			if (hayform) {
				var $error  = $('[name=' +error+']', form);
			} else {
				var $error  = $('[name=' +error+']', '[instance="'+parentInstance+'"]');
			}

			$error.attr('valid', false);

			if ($error.attr('novalidar') == 'true')
				$error.attr('valid', true);

			obtengo = false;
		}
	}

	Histrix.buscando = false;
	// alert(originalRequest.responseXML);
	// Si la ayuda devuelve mas de un registro
	var xmlaux = xml.getElementsByTagName('xmlaux').item(0);
	if (xmlaux != null) {
		var id = xmlaux.getAttribute('id').toString();
		var instance = xmlaux.getAttribute('instance').toString();
		obtengo = false;
		// Abro otra ventana
		TablaAyuda(id, instance);
		return;
	}

	if ($resultado.attr('campoOrigen')) {
		var campoOrigen = $resultado.attr('campoOrigen');
	}

	if (obtengo == true) {
	
		$(source).attr('valid', 'true');

		xmldatos = $resultado.attr('xmlorigen');

		// if (resultado.hasAttribute( 'xmlpadre' )){
		// var xmlpadre = resultado.getAttribute( 'xmlpadre' ).toString();
		// }
		if ($resultado.attr('modificar') !== undefined)
			var mod = $resultado.attr('modificar');

		if ($resultado.attr('noactboton') !== undefined)
			var noactboton = $resultado.attr('noactboton');

//		loger($resultado);
		if (xmldatos){
			var idxmldatos = xmldatos.replace(".", '_');
			var xmlOriginal = xmldatos.replace("_xml", '.xml');
		}
//		

		misforms = $('form[instance="'+parentInstance+'"], div[instance="'+parentInstance+'"]' );

		if (misforms.length == 0) misforms = $('form[instance="'+instance+'"], DIV[instance="'+instance+'"]');

/*		var misforms = []; */
		if (misforms.length > 0 ) {
			hayform = true;
		}
/*
		if (form.length > 0) {
			hayform = true;
			misforms[0] = form[0];
		}

		// TEST HELPS IN FILTERS!!!
		// 
		// Modificacion para que funcione las ayudas con los filtros
		// No me traia el detalle sino
		if ($('#FForm' + idxmldatos)[0]) {
			hayform = true;
			form = $('#FForm' + idxmldatos);
			misforms[1] = form[0];
		}
*/
		var campo;
		var destino;
		var result = xml.getElementsByTagName('campo');

		for ( var i = 0; i < result.length; i++) {
			campo = result.item(i);

			if (campo.getAttribute('destino') != undefined) {
				destino = campo.getAttribute('destino').toString();
				var valorTemp = $(campo).text();
				// campo.childNodes.item(0).firstChild.data;
				vars += destino + "=" + encodeURIComponent(valorTemp) + '&';
			}
		}

		if (mod != 'false') {
			// loger('lleno Campos desde XML');
			/*
			var xmlOrig = $(form).attr('original');
			var xmlorigvar = '';
			if (xmlOrig)
				xmlorigvar = '&xmlOrig=' + xmlOrig;
			*/
			$.post("setData.php?_show=false&instance=" + instance + "&modificar=true" , vars);

		}

		// force valid origin field
		// in case no detail is provided

		//TODO: USE SOURCE OBJECT
		if (obtengo == true) {

			$('input[name="' + campoOrigen + '"]', form).attr('valid', true);

			// DISABLE
			// key fields 
				
			$("input.clave", form).attr('disabled', 'disabled');
		}




		result = xml.getElementsByTagName('campo');

		for (i = 0; i < result.length; i++) {
			campo = result.item(i);

			// var idcampo = campo.getAttribute( 'id' ).toString();
			if (campo.getAttribute('obj') != undefined) {
				var xmlrefrescar = campo.getAttribute('obj').toString();
				if (xmlrefrescar != '') {
					var idxmlrefrescar = xmlrefrescar.replace(".", '_');
					var innerInstance = $('#' + idxmlrefrescar+ ' [instance]').attr('instance');
					if (innerInstance){
						$('#' + idxmlrefrescar, '[instance="'+instance+'"]').load(
								"refrescaCampo.php?xmldatos=" + xmlrefrescar
										+ '&instance=' + innerInstance
										+ '&xmlpadre=' + xmldatos, {'__winid':Histrix.uniqid});

					}
				}
			}

			// Search and update input field
			if (campo.getAttribute('destino') != undefined) {
				destino = campo.getAttribute('destino').toString();
				var valor = $(campo).text();
				// campo.childNodes.item(0).firstChild.data;
				vars += destino + "=" + encodeURIComponent(valor) + '&';

				if (hayform) {

					// cicle all forms???
					// see if this can be removed

					for ( var f = 0; f < misforms.length; f++) {
						var myform = misforms[f];
						// var myformID = $(myform).attr('id');
loger(destino);
loger(myform);
						if (myform == undefined)
							continue;

						$(' :input[name='+campoOrigen+']', myform).attr('valid', true);

						$(' :input[name='+destino+']', myform).each(
										function () {

											tempobj = $(this);
loger(destino);
loger(tempobj);
												tempobj.attr('valid', true);

												if (tempobj[0].type == "radio") {
													if (valor == tempobj.val())
														tempobj[0].checked = true;
													else
														tempobj[0].checked = false;

												} else
													tempobj.val(valor);

												if (tempobj[0].type == "checkbox") {
													if (valor == 1)
														tempobj[0].checked = true;
													else
														tempobj[0].checked = false
												}
												
												if (tempobj.attr('image') != undefined) {
													tempobj.change();
												}

												if (tempobj.attr('onchange') != undefined
														&& tempobj.attr('name') != campoOrigen) {
													// sino se vuelve recursivo
													// e infinito
													var onc = tempobj.attr(
															'onchange')
															.toString();
													if (onc.indexOf('actuali') != -1 ||  onc.indexOf('activate') != -1 ) {
														tempobj.change();
													}
												}

												if (tempobj.attr('autoing')
														&& valor != '') {
													formname = 'Form' + xmldatos;
													if (tempobj.attr('innerform') != undefined)
														formname = tempobj.attr('innerform');

													xmlfile = xmlOriginal;

													if (tempobj.attr('innerxml') != undefined)
														xmlfile = tempobj.attr('innerxml');

													var $divinstance = tempobj.closest('[instance]');

													if ($divinstance.length != 0)
														xmlinstance = $divinstance.attr('instance');

													autosube = true;

												}

										});
					}
					
					
				} else {
					// $(destino).value = valor;
					// Without Form
				}
			} // end if destino
		} // end for each result

		llenoCabecera($(myform).attr('id'), xmldatos, xmldatos, tempobj);

		// update buttons //
		/*
		var myform2 = '';
		for ( var f2 = 0; f2 < misforms.length; f2++) {
			myform2 = misforms[f2];
			if (myform2 == undefined)
				continue;
			var Fid = $(myform2).attr('id');
	*/

			var obj = $('#tformfoot_' + idxmldatos + ' .filabotones button:visible');
			
			obj.each(function (index, elem) {
				if (elem.name == 'Grabar' && noactboton != 'true') {


					$(elem).html(
							'<img src="../img/filesave.png" alt="'
									+ Histrix.i18n['save'] + '"  />'
									+ Histrix.i18n['save']).attr("accion",
							'update');
				}
				
				if (elem.name == 'delete' && noactboton != 'true') {
					elem.removeAttribute("disabled");
				}
			});
	//	}

		if (autosube == true && obtengo == true) {
			grabaABM(formname, 'insert', xmlfile, xmlfile, true, 'Form'
					+ xmlOriginal, {
				'instance' : xmlinstance
			});
		}
	}
}

function TempTableSort(lnk) {

	var td = lnk.parentNode;
	var table = getParent(td, 'TABLE');
	var tblLen = table.rows.length;
	if (tblLen <= 1)
		return;
	var xmldatos = table.id.substring(12);
	xmldatos = xmldatos.replace("_xml", '.xml');
	$table = $(table);
	var xml = $table.attr('xml');
	var instance = $table.attr('instance');
	if (xml != undefined) {
		xmldatos = xml;
	}

	var xmlOrig = $table.attr('xmlOrig');

	var newOrder = []; // 80' music is everywhere
	$('#' + table.id + ' > tbody  > tr[o]').each(function (row) {
		var $this = $(this);
		newOrder[row] = $this.attr('o');
		$this.attr('o', row);
	})
	if (xmldatos != '') {
		var vars = "&newOrder=" + newOrder;
		$.post("setData.php?_show=false&instance=" + instance, vars);
	}
}

function TablaAyuda(xmlayuda, instance) {
	loger('Tabla Ayuda jQuery' + xmlayuda);
	var xmlprincipal = xmlayuda.substring(5);
	// var xmlprincipal = xmlayuda.replace("_aux_", '');

	var idxmlprincipal = xmlprincipal.replace(".", '_');
	var idxmlayuda = xmlayuda.replace(".", '_');

	if ($('#' + idxmlayuda)[0] == undefined) {

		var supra = $('#' + idxmlprincipal);
		if (supra[0] == undefined)
			supra = $('#DIVFORM' + idxmlprincipal);

		var newdiv = jQuery('<div></div>').attr("id", 'HLP' + idxmlayuda)
				.addClass('ventint').css({
					top : '50px',
					left : '50px'
				});

		supra.append(newdiv);

		var titulo = 'Seleccione el Registro deseado';
		var contenidohtml = barraDrag('HLP' + idxmlayuda, titulo);
		var divresize = barrawin('HLP' + idxmlayuda);
		newdiv.html(contenidohtml);

		var newdiv2 = jQuery('<div></div>').addClass('contewin').attr('id',
				'TEMPORAL' + idxmlayuda);
		Histrix.loadingMsg(newdiv2.attr('id'), titulo);

		newdiv.append(newdiv2);
		newdiv.append(divresize);
		$('#HLP' + idxmlayuda).draggable({
			handle : '#dragbarHLP' + idxmlayuda
		});

		newdiv2.load("process.php?instance=" + instance + "&xmldatos="
				+ xmlayuda + "&divcont=HLP" + xmlayuda + '&forzar=consulta', {'__winid':Histrix.uniqid});

	}
}

function filtrar(campo, valor, operador, tabla, xmldatos, objeto, xmlOrig) {

	var idtabla = tabla.replace(".", '_');
	var instance = $(objeto).closest('FORM').attr('instance');

	objeto.options[objeto.selectedIndex].defaultSelected = true;
	if (valor != '') {
		Histrix.loadingMsg(idtabla, 'Filtrando')
		$('#' + idtabla).load(
				"process.php?xmldatos=" + xmldatos + "&xmlOrig=" + xmlOrig
						+ "&filtro=" + campo + "&valor=" + valor + "&instance="
						+ instance + "&operador=" + operador
						+ "&reemplazo=reemplazo", {'__winid':Histrix.uniqid});
	}
}

function reloadImg(img) {
	var now = new Date();
	var src = $('#IMG' + img).attr('src');
	$('#IMG' + img).attr('src', src + '&' + now.getTime());
}

function softClearForm(elem) {
	$(elem).parents('form:eq(0)').find(
			'input:visible:enabled,textarea,select:visible:enabled').val('')
			.removeAttr('checked').removeAttr('selected');

}

function llenoTablaIng(IDformu, IDtabla) {
	var formu = $('#' + IDformu)[0];
	var objTabla = $('#' + IDtabla);
	var fila0 = objTabla.insertRow(objTabla.rows.length - 2);
	var fl = formu.length;
	var ultima = fl;
	var celda;
	for (i = 0; i < fl; i++) {
		var tempobj = formu.elements[i];
		if (tempobj.type != "button" && tempobj.type != "reset") {
			celda = fila0.insertCell(i);
			if (document.all) {
				celda.innerText = tempobj.value;
			} else {
				celda.textContent = tempobj.value;
			}
			ultima = i;
		}
	}

	ultima++;
	celda = fila0.insertCell(ultima);
	var js = '<img src="../img/remove.png" title="Borrar" '
			+ 'onclick="this.parentNode.parentNode.parentNode.deleteRow(this.parentNode.parentNode.rowIndex);" >';
	celda.innerHTML = js;
	objTabla.Height = 300;

}

function setGrafico(form, xmldatos, idgraf, instance) {
	var idxmldatos = xmldatos.replace(".", '_');

	if (form != null) {
		var vars = ObtengoValoresForm(form, true, 'obj');
	}
	if (idgraf != null) {
		var action = "&accion=update&idgraf=" + idgraf;
	} else
		action = "&accion=add";

	Histrix.loadingMsg('GRAFINT' + idxmldatos, 'Grafico');
	vars['__winid'] = Histrix.uniqid;
	$('#GRAFINT' + idxmldatos).load(
			"AbmGraf.php?instance="+ instance +"&xmldatos=" + xmldatos + action, vars);
}

function addGraficos(xmldatos, idgraf) {
	var tempX = 10;
	var tempY = 10;
	var titulo = 'Modificacion de Graficos';

	var idxmldatos = xmldatos.replace(".", '_');

	var supra = $('#DIV' + idxmldatos)[0];
	if (!supra) {
		supra = $('#Show' + idxmldatos)[0];
	}
	var newdiv = document.createElement('div');
	$(newdiv).attr('id', 'GRAF' + idxmldatos).addClass('ventint').css({
		width : '60%',
		top : tempY + 'px',
		left : tempX + 'px'
	});

	supra.appendChild(newdiv);

	var getidgraf = '';
	if (idgraf != undefined)
		getidgraf = '&idgraf=' + idgraf;


	var instance = $('[instance]'.supra).attr('instance');

	var contenido = 'AbmGraf.php?instance='+ instance +'&xmldatos=' + xmldatos + getidgraf;

	var contenidohtml = barraDrag('GRAF' + idxmldatos, titulo);
	var divresize = barrawin('GRAF' + idxmldatos);

	newdiv.innerHTML = contenidohtml;
	var newdiv2 = document.createElement('div');

	$(newdiv2).addClass('contewin').attr('id', 'GRAFINT' + idxmldatos);

	Histrix.loadingMsg(newdiv2.id, titulo);

	newdiv.appendChild(newdiv2);
	newdiv.appendChild(divresize);

	$('#GRAF' + idxmldatos).draggable({
		handle : '#dragbarGRAF' + idxmldatos
	});
	$(newdiv2).load(contenido, {'__winid':Histrix.uniqid});
}

function deleterow(numrow, contenedor, xmldatos, xmlOrig, button) {
	var idxmldatos = xmldatos.replace(".", '_');
	Histrix.showMsg('Borrando Fila');
	if (contenedor == '')
		contenedor = 'Form' + idxmldatos;
	if (!$('#' + contenedor)[0])
		contenedor = idxmldatos;

	var instance = $(button).closest('table').attr('instance');
	if (instance == undefined)
		instance = $(button).closest('form').attr('instance');

	$('#' + contenedor).load(
			"process.php?instance=" + instance + "&xmldatos=" + xmldatos
					+ "&accion=delete" + "&xmlOrig=" + xmlOrig, 
					{
						rowaborrar : numrow,
						'__winid':Histrix.uniqid
					}, 
			function (){
				Histrix.hideMsg();
				Histrix.calculoAlturas(idxmldatos);
			});
		
}



// Input Default Value acts like ghosts :)
function ghost(obj, ev) {
	// loger('ghost jQuery'+ev.type);
	var inp = $(obj);
	var defaultValue = inp.prop('defaultValue');
	if (ev.type == 'blur') {
		if (inp.val() == '')
			inp.val(defaultValue);
	}
	if (ev.type == 'focus') {
		if (inp.val() == defaultValue)
			inp.val('');
	}
}

function validateProcess(validations, messages, validationOptional, instance) {
	// loger('valido');
	if (validations != undefined) {

		var validationsLength = validations.length;

		for ( var k = 0; k < validationsLength; k++) {
			/*
			var formula = Histrix.evaluate(validations[k], 'Form'
					+ idxmldatos, $('#'+ idxmldatos));    
			*/

			var formula = Histrix.evaluate(validations[k], $('[instance="'+instance+'"]'));    

			/*
			// add header to form validation
			if (idxmlcabecera != undefined){
				formula = Histrix.evaluate(formula, 'Form'
					+ idxmlcabecera);                                				
			}
			*/
			valor = false;
			try {
				 loger('validations : '+ formula);
				var valor = eval(formula);
			} catch (ex) {
				 loger(formula + ex);

			}

			if (valor) {
				if (validationOptional[k] == 'true') {
					if (!confirm(messages[k])) {
						return false;
					}
				} else {
					jQuery('<div><p><span class="ui-icon ui-icon-alert" style="float: left; margin: 0 7px 20px 0;"></span>'+messages[k]+'</p></div>').dialog(
					    {modal:true,
					    resizable: false,
					    buttons: {
				                Ok: function() {
			            		        $( this ).dialog( "close" );
					            }
				                }
					    }

					);
					//alert(messages[k]);

					return false;

				}
			}
		}
	}





	return true;
}
/**
 * Process button get Data from Internal forms and grids and send results to
 * process.php
 */

function procesar(
// xmldatos,
// xmlcabecera,
// button,
// mensajeconf,
// dir,
// cierraproceso,
// arrayGrillas,
// validations,
// messages,
// validationOptional,
// div_destino ,

	postOptions // this will replace all above
	) {
	// Default windows Options
	var defaultOptions = {
		xmlOrig : '',
		cierraprocesoCondition : '',
		cierraprocesoDir : ''
	}
	var options 	= $.extend(defaultOptions, postOptions);

	var xmldatos 			= options.xmldatos;
	var dir 				= options.dir;
	var xmlcabecera 		= options.xmlcabecera;
	var button 				= options.button;
	var mensajeconf 		= options.mensajeconf;
	var cierraproceso 		= options.cierraproceso;
	var cierraprocesoCondition = options.cierraprocesoCondition;
	var arrayGrillas 		= options.arrayGrillas;
	var instanceGrids 		= options.instanceGrids;

	var validations 		= options.validations;
	var messages 			= options.messages;
	var validationOptional 	= options.validationOptional;

	var div_destino = options.div_destino

	var myregexp = new RegExp('.xml', "g");
	var idxmldatos = xmldatos.replace(".", '_');
	var idxmlcabecera = xmlcabecera.replace(".", '_');

	var myvars = {};

	// perform logical validation
	var validated = validateProcess(validations, messages, validationOptional, options.instance);

	if (!validated)
		return validated;

	try {
		$('[internal_class="simpleditor"]').wysiwyg("setContent");
	} catch (ex) {}


	//
	/*
	 * var grilla=tabla.parentNode.parentNode.id; // rehacer loger('calculo
	 * Total'); var instance = $(tabla).closest('table').attr('instance');
	 * 
	 * $.post("setData.php?xmldatos="+grilla+'&instance='+instance+'&xmlOrig='+xmlParent+"&actualizoTabla=true", {
	 * mijson:miJSON });
	 */
	//
	var procesar = true;
	if (mensajeconf != '')
		var conf = confirmacion(mensajeconf);
	if (conf == false) {
		Histrix.hideMsg();
		return conf;
	}

	var formu;
	var chequeo;

	var $dialog = false;

	var xmlForm = $('#Form' + idxmldatos);

	if (button != undefined) {

		var $xmlForm = $(button).closest('#Form' + idxmldatos);
		if ($xmlForm.length != 0)
			xmlForm = $xmlForm;
	}

	var closecondition = true;
	if (cierraprocesoCondition != '') {

		var formula = Histrix.evaluate(cierraprocesoCondition, xmlForm);

		var closecondition = eval(formula);
	}

	if (cierraproceso != '' && closecondition) {
		formu = $('#Form' + idxmlcabecera)[0];

		if (formu) {
			if ($wf2.isInitialized) {
				if ($wf2)
					$wf2.onDOMContentLoaded();
			}

			chequeo = Histrix.checkForm(formu);
				
			if (!chequeo){
				formu.checkValidity();
				return false;
			}
		}

		// // VER si hace falta!
		// / controlar repeticiones
		if (xmlForm[0]) {
			if (xmlForm.attr('tipo') == 'fichaing') {
				var instance = xmlForm.attr('instance');
				$.post("setData.php?_show=false&instance=" + instance, ObtengoValoresForm(
								xmlForm[0], true, 'obj'));
			}
		}
		// // VER si hace falta! END                                     
		procesar = false;
		var customDir = ( options.cierraprocesoDir == '')? dir : options.cierraprocesoDir;

		Histrix.ventInt(xmldatos, cierraproceso, '&cierraproceso=true&dir=' + customDir
				+ '&parentInstance=' + options.instance, "", {
			'modal' : true,
			'parentInstance':options.instance
		});
	}

	if (procesar) {
		// Disable Button to prevent DOUBLE PRESS
		if (button)
			button.disabled = true;

		var vars = "";

		formu = $('#Form' + idxmlcabecera)[0];

		// Para procesar formularios tipo fichaing
		if (xmlForm[0]) {
			if (xmlForm.attr('tipo') == 'fichaing') {

				myvars = ObtengoValoresForm(xmlForm[0], true, 'obj');
			}
		}

		if (formu) {
		    if ($wf2.isInitialized) {
			if ($wf2)
			    $wf2.onDOMContentLoaded();
		    }

			if (formu.checkValidity()) {

				chequeo = Histrix.checkForm(formu);

				if (!chequeo) {
					loger('ERROR');
					return false;
				}
			} else {
				return false;
			}

			var fl = formu.length;
			var separator = '';
			for (i = 0; i < fl; i++) {
				var tempobj = formu.elements[i];
				var valortmp = tempobj.value;
				if (tempobj.type == "checkbox") {
					if (tempobj.checked)
						valortmp = 1;
					else
						valortmp = 0;
				}

				if (tempobj.type != "button" && tempobj.type != "reset"
						&& tempobj.name != '') {
					// vars += tempobj.name + '="' + valortmp + '"&';
					// Si le saco el encode se rompe la serializacion de
					// textareas (orden de compra por ejemplo)
					// verificar

					// vars += separator+tempobj.name + '="' +
					// encodeURIComponent(valortmp) + '"';
					myvars[tempobj.name] = encodeURIComponent(valortmp);
					// separator='&';
				}
			}
		}

		$dialog = Histrix.showMsg('Procesando.. ' + idxmldatos, true);

		var destino = $('#' + idxmldatos);
		if (!destino[0]) {
			destino = $('#DIV' + idxmldatos);
		}

		if (div_destino) {
			destino = $('#' + div_destino);

		}
		if (destino.length == 0) {

			destino = $(button).closest('#DIVFORM' + idxmldatos);

			// destino = $('#DIVFORM'+idxmldatos);
		}

		var xmlOrig = $(xmlForm).attr('original');
		var xmlorigvar = '';
		if (xmlOrig) {
			xmlorigvar = '&xmlOrig=' + xmlOrig;
		} else {
			xmlorigvar = '&xmlOrig=' + options.xmlOrig;
		}
		Histrix.loadingMsg(destino.id, 'Procesando Datos');

		if (options.instance != '') {
			xmlorigvar += '&instance=' + options.instance;
		}

		if (arrayGrillas != '' && arrayGrillas != undefined) {
			var grillas = arrayGrillas.length;
			var procesadas = 0;
			// Recorro las grillas editables y actualizo sus respectivos
			// Contedores
			for ( var i = 0; i < grillas; i++) {
				var grilla = arrayGrillas[i];
				var miJSON = Histrix.updateGridById(grilla);
				var instance = instanceGrids[i];

				loger('Procesar' + instance);
				$.post("setData.php?_show=false&xmldatossub=" + xmlcabecera
						+ "&actualizoTabla=true&instance=" + instance, {
					mijson : miJSON
				}, function () {
					procesadas++;
					if (procesadas == grillas) {
						//loger('PROCESO');
						myvars['__winid'] =Histrix.uniqid;
	/*
						$(destino).load(
								"process.php?xmldatos=" + xmldatos 
								+ "&accion=procesar" + xmlorigvar,
								myvars, 
								function (responseText, textStatus){

								 	Histrix.hideMsg();
								 	$dialog.dialog("close");

									if (textStatus == 'error'){
										Histrix.alerta(responseText, undefined, undefined, "Error de Grabacion");
										if (button)
											button.disabled = false;
										return false;
									}


								});
				    */

			$.ajax({type:"POST",
			    url: "process.php?xmldatos=" + xmldatos + "&accion=procesar"
				    + xmlorigvar, 
			    data: myvars, 
			    complete: function(){
				    Histrix.hideMsg();
			    },
			    error: function(jqXHR){
				Histrix.alerta(jqXHR.responseText, undefined, undefined, "Error de Grabación!");
				Histrix.hideMsg();
				$(button).button('enable');
				if (button)
					button.disabled = false;
				return false;

			    },
			    success: function(responseText, textStatus, jqXHR ){
				
					$(destino).html(responseText);
							        
				        if ($dialog)
				    	$dialog.remove();
				    
					Histrix.hideMsg();
					$dialog.dialog("close");

				    $(button).button({enhanced:true});
				    $(button).button('enable');									

			    }
			});

					}

				});
			}
		    Histrix.hideMsg();
//		    $dialog.dialog("close");			
		} else {
			myvars['__winid'] =Histrix.uniqid;
/*
			$(destino).load(
					"process.php?xmldatos=" + xmldatos + "&accion=procesar" 
							+ xmlorigvar, myvars, 
							function (responseText, textStatus){
									if (textStatus == 'error'){
										Histrix.alerta(responseText, undefined, undefined, "Error de Grabacion");
										if (button)
											button.disabled = false;
										return false;
									}

								 	Histrix.hideMsg();
								 	$dialog.dialog("close");
							});
		    */

			// 13 replace with post set instance data
			$.ajax({type:"POST",
			    url: "process.php?xmldatos=" + xmldatos + "&accion=procesar"
				    + xmlorigvar, 
			    data: myvars, 
			    complete: function(){
				    Histrix.hideMsg();
			    },
			    error: function(jqXHR){
				Histrix.alerta(jqXHR.responseText, undefined, undefined, "Error de Grabación..");
				Histrix.hideMsg();
				$(button).button('enable');
				if (button)
					button.disabled = false;
				return false;

			    },
			    success: function(responseText, textStatus, jqXHR ){
				
					$(destino).html(responseText);
							        
				        if ($dialog)
				    	$dialog.remove();
				    
					Histrix.hideMsg();
					$dialog.dialog("close");

				    $(button).button({enhanced:true});
				    $(button).button('enable');									

			    }
			});

		}

	} else {
		/* tengo que recargar la cabecera */
		Histrix.hideMsg();
		if ($dialog)
			$dialog.dialog("close");
		return false;
	}
	//Histrix.hideMsg();
	// TESTING
	Histrix.sysmon();

}

function grabaABM(IDformu, accion, contenedor, xmldatos, validate, formuCab,
		options) {
	loger('grabaAbm: ' + IDformu + accion + contenedor + xmldatos + validate
			+ formuCab);

	var defaultOptions = {
		xmlpadre : false,
		processEvent : true,
		instance : '',
		parentInstance : '',
		button : null,
		clickedbutton : null,
	}
	var postOptions = $.extend(defaultOptions, options);

	IDformu = IDformu.replace(".", '_');
	// var idxmlcabecera = xmlcabecera.replace(myregexp, '_xml');
	if (xmldatos == '')
		return false;
	if (validate == undefined)
		validate = true;
	var idxmldatos = xmldatos.replace(".",'_');




	/* add validations to method */
	var validations 		= postOptions.validations;
	var messages    		= postOptions.messages;
	var validationOptional  = postOptions.validationOptional;
	// perform logical validation
	var validated = validateProcess(validations, messages, validationOptional , postOptions.instance);
	if (!validated)
		return false;


	var chequeo = true;
	// for Nested inner forms
	if (formuCab != undefined && IDformu != formuCab) {
		formuCab = formuCab.replace(".", '_');

		var innerForm = $('#'+formuCab, '[instance="'+postOptions.parentInstance+'"]')[0];
		if (innerForm) {

			if ($wf2.isInitialized) {
				if ($wf2)
					$wf2.onDOMContentLoaded();
			}

			if (innerForm.checkValidity()) {

				chequeo = Histrix.checkForm(innerForm);
			} else {
				loger('no cab checkValidity');
				return false;
			}
		}
	}
	
	var inner  = false;
	$formu = $('#' + IDformu , '[instance="'+postOptions.instance+'"]').not('[tipo=filter]');
	if ($formu.length == 0){
		$formu = $('form[instance="'+postOptions.instance+'"]').not('[tipo=filter]');
	}
	
	if ($formu.length == 0){
		$formu = $('.form[instance="'+postOptions.instance+'"]').not('[tipo=filter]');
		inner = true;
	}
	if ($formu.length == 0){
		$formu = $('[instance="'+postOptions.instance+'"]').not('[tipo=filter]');
	}
	
/*
	// if i cant find main form
	if ($formu.length == 0) {
		$formu = $('#' + idxmldatos).closest('form');
	}
*/
	var formu = $formu.get(0);


	if (formu && validate) {
        loger('intento validar formulario: '+ formu.id);
		if ($wf2.isInitialized) {
			if ($wf2)
				$wf2.onDOMContentLoaded();
		}
    
		
		chequeo = Histrix.checkForm(formu);

		if (chequeo) {
			// formulario ES Valido
		} else {
		    // formulario NO es Valido
		    loger('no checkValidity');
		    formu.checkValidity()
		    return false;
		}
	}

	var parentForm =  $(formu).attr('original');
	if (parentForm != undefined)
        	parentForm 	= parentForm.replace(".", '_');
        var $parent =  $('#Form'+parentForm+'[instance="'+postOptions.parentInstance+'"]');


	if ($parent[0] != undefined ) {


    	if ($parent[0].checkValidity()){

	    }
	    else {
			loger('NO VALIDO PARENT FORM');
			return false;
	    }

	
	}

	// var i = 0;
	// var fl = 0;
	if (accion == 'delete')
		var conf = confirmacion(Histrix.i18n['deleteQuestion']);
	if (conf == false) {
		Histrix.hideMsg();
		return conf;
	}
	var vars = "";
	// Si no tengo formulario intento hacerlo recorriendo los elementos
	if (!formu) {
		//loger('sin formu');
		var div = $('#TR' + IDformu);

		if (postOptions.button){
			var $outer =  $(postOptions.button).closest('form');
			div = $('#TR' + IDformu, $outer);
		}

		var vars = Histrix.getNodeData(div[0]);

	} else {
		//loger('con formu');
		if (validate != false) {
		//	chequeo = Histrix.checkForm(formu);
		}

		if (chequeo == true) {
			vars = ObtengoValoresForm(formu, true, 'obj', inner);
		} else {
			loger('no pasa checkeo');
		}
	}

	if (contenedor == '')
		contenedor = xmldatos;

	if (chequeo == true) {

		Histrix.showMsg(accion);
		
		contenedor = contenedor.replace(".", '_');

		var row = Histrix.getSelectedRow(contenedor)[0];
		var rowIndex = 0;
		if (row)
			rowIndex = Histrix.getSelectedRow(contenedor)[0].rowIndex;

		// TODO
		// MODIFICAR ESTE METODO PARA QUE LOS UPDATES DE RENGLONES SOLAMENTE
		// CAMBIEN 1 RENGLON VIA
		// JAVASCRIPT

		if (!postOptions.clickedbutton)
			postOptions.clickedbutton = postOptions.button;

		var OBJ = $(postOptions.clickedbutton).closest('#' + contenedor);


		if ($outer != undefined){
		    OBJ = $('#' + idxmldatos, $outer);
		} 
		if (OBJ.hasClass('detalle')){

			OBJ = $('#'+idxmldatos);
		}

		// type abm
		if (OBJ.length == 0){
			OBJ = $('div[instance="'+postOptions.instance+'"]');
		}
		

		if (OBJ[0]) {
			// nada
		} else {
			// otro obj;
			IDformu = IDformu.replace(".", '_');
			OBJ = $('#' + IDformu+'[instance="'+postOptions.instance+'"]');
		}

		var xmlOrig = $(formu).attr('original');
		var xmlorigvar = '';

		if (postOptions.xmlpadre != false)
			xmlOrig = postOptions.xmlpadre;

		if (xmlOrig)
			xmlorigvar = '&xmlOrig=' + xmlOrig;
		// else loger('no encontre en '+xmldatos);
		if (postOptions.instance != '') {
			xmlorigvar += '&instance=' + postOptions.instance;
		}

		// disable button to prevent event duplication
		
		$(postOptions.clickedbutton ).attr('disabled' ,'disabled');
		

		vars['__winid'] =Histrix.uniqid;
		OBJ.load(
						"process.php?_pe="+postOptions.processEvent+"&accion=" + accion + "&xmldatos="
								+ xmldatos + xmlorigvar + "&delform=" + IDformu,
						vars,
						function (responseText, textStatus) {

							Histrix.hideMsg();

							$(postOptions.clickedbutton).removeAttr('disabled');




							if (textStatus == 'error'){
								// after error status
								Histrix.alerta(responseText, undefined, undefined, "Error de Grabacion");

								// Clear Form Anyway?
								if (accion == 'delete') {
									Histrix.clearForm(IDformu, true);
								}

							}
							else {

								// after success status
								if (accion == 'delete') {
									Histrix.clearForm(IDformu, true);
								} else {
									/*
									if (formu) {

										var obj = $('#' + IDformu + ' button[idform='+idxmldatos+']');
										obj.each(function () {
											var elem = $(this);
											
											if (elem.attr("accion") == 'insert' && accion != 'refresh') {
												
												elem.html(
														'<img src="../img/filesave.png" alt="'
																+ Histrix.i18n['save'] + '"  />'
																+ Histrix.i18n['save']).attr("accion",
														'update');
																
											}
											
										});
									}
									*/
								}
															
								Histrix.calculoAlturas(xmldatos);
								var tbody = $('#tbody_' + contenedor)[0];
								if (tbody) {
									var table = tbody.parentNode;
									positionRow(table, rowIndex, accion);
								}
								if ($(formu).attr('tipo') != 'ficha' && ($(formu).attr('tipo') == 'abm'
												|| $(formu).attr('tipo') == 'abm-mini' 
												|| postOptions.button != undefined)) {

									Histrix.clearForm(
													IDformu,
													true,
													postOptions.button,
													{
														innerForm : postOptions.button.idform
													});
								}
							}

						});
	} else {
		return false;
	}



	$('#DIVFORM' + contenedor + '.singleForm').each(function () {
		$(this).slideUp();
	});

	return true;
}

function IsNumeric(valor) {
	var log = valor.length;
	var sw = "S";
	var v1 = 0;
	var v2 = 0;
	for (x = 0; x < log; x++) {
		v1 = valor.substr(x, 1); 
		v2 = parseInt(v1);
		// Compruebo si es un valor numérico
		if (isNaN(v2)) {
			sw = "N";
		}
		// Chequeo caracteres permitidos
		if (v1 == '.' || v1 == ',' || v1 == '-') {
			sw = "S";
		}
	}
	if (sw == "S") {
		return true;
	} else {
		return false;
	}
}

function formateafecha(fecha) {
	if (fecha) {
		var longs = fecha.length;
		var dia;
		var mes;
		var ano;
		var d1;
		var d2;
		dia = fecha.substr(0, 2);

		switch (longs) {
		case 1:
			if (IsNumeric(dia) == false)
				fecha = "";
			break;
		case 2:
			d1 = dia.substr(0, 1);
			d2 = dia.substr(1, 2);
			if (d2 == '/' || d2 == '-') {
				dia = '0' + d1;
			}

			if ((IsNumeric(dia) == true) && (dia <= 31) && (dia != "00")) {
				fecha = dia + "/";
			}
			break;
		case 4:
			mes = fecha.substr(3, 1);
			if (IsNumeric(mes) == true) {
				fecha = dia + "/" + mes
			}
			break;
		case 5:
			mes = fecha.substr(3, 2);
			d1 = mes.substr(0, 1);
			d2 = mes.substr(1, 2);
			if (d2 == '/' || d2 == '-') {
				mes = '0' + d1;
			}

			if ((IsNumeric(dia) == true) && (dia <= 31) && (dia != "00")) {
				fecha = dia + "/" + mes + "/";
			}
			break;
		}
		if (longs > 10)
			fecha = fecha.substr(0, 10);
	}

	return (fecha);
}

function showImage(obj, img, event, supra1, filename, uid) {
	// loger('ShowImg jQuery');
	var supra = Histrix.Supracontenido;
	if (supra1 != undefined) {
		supra = $('#' + supra1);
	}
	var newdiv = jQuery('<div ></div>').attr('id', uid).css('width', '270px');

	var html;
	if (obj) {

		var pos = Histrix.getPosition(event);
		var tempX = pos.x;
		var tempY = pos.y;
		var h = document.body.offsetHeight;
		var w = document.body.offsetWidth;

		if (tempY + 250 > h)
			tempY = h - 500;
		if (tempX + 250 > w)
			tempX = tempX - 300;

		newdiv.css({
			top : (tempY - obj.clientHeight) + 'px',
			left : (tempX + 3) + 'px'
		}).addClass('polaroid').click(function () {
			cerrarVent(uid);
		});

		html = '<img art="' + filename
				+ '" src="thumb.php?ancho=250&alto=500&url=' + img
				+ '" align="middle">';
		html += '<br><span>' + decodeURIComponent(filename) + '</span>';
	} else {
		newdiv.addClass('ayudaint').draggable({
			handle : '#dragbar' + newdiv.attr('id')
		});
		html = '<div class="popalerta"  ><p>' + mensaje
				+ '</p><input type="button" value="cerrar" '
				+ 'name="cerrar" onClick="cerrarVent(\'' + uid + '\');"></div>';
	}
	newdiv.html(html);
	$(supra).append(newdiv);
}

function getTR(node) {
	while (node && node.nodeName != "TR") {
		node = node.parentNode;
	}
	return node;
}


function isScrolledIntoView(elem) {
	var docViewTop = $(window).scrollTop();
	var docViewBottom = docViewTop + $(window).height();

	var elemTop = $(elem).offset().top;
	var elemBottom = elemTop + $(elem).height();

	return ((elemBottom >= docViewTop) && (elemTop <= docViewBottom));
}

function positionRow(table, rowNumber, accion) {
	var myRow = $(table)[0].rows[rowNumber];
	Histrix.selectRow(myRow);

	// if (!isScrolledIntoView(myRow))
	$(myRow).closest('div.tablewrapper').scrollTo($(myRow));

	if (accion == 'update')
		$(myRow).pulse({
			backgroundColors : [ 'yellow', 'orange' ],
			runLength : 1,
			speed : 500
		});
}

function colorear(campo) {
	var valor = campo.value.substr(0, 6);
	$(campo).css('backgroundColor', "#" + valor);
}

// File Picker modification for FCK Editor v2.0 - www.fckeditor.net
// by: Pete Forde <pete@unspace.ca> @ Unspace Interactive
var urlobj;

/*
function BrowseServer2Old(obj, Type, url, input, inpObj) {
	var btn = $(inpObj);
	if (btn.getAttribute('pathObj') != undefined) {
		url = url + $(btn.getAttribute('pathObj')).value + '/';
	}
	OpenServerBrowser('../lib/filemanager/filemanager.php?basedir=' + url
			+ '&input=' + input, screen.width * 0.5, screen.height * 0.5, input);
}
*/
function BrowseServer2(obj, Type, url, input, inpObj) {
	var btn = $(inpObj);
	if (btn.attr('pathobj') != undefined) {
		//var $field = $('[name="' + btn.attr('pathobj') + '"]' , btn.closest('form'));
		$field = $('#'+btn.attr('pathobj'));
		url = url +  $field.val() + '/';
	}
	var maxWidth = $(inpObj).closest('[maxWidth]').attr('maxWidth');
	OpenServerBrowser('fileManager.php?access=rwd&basedir=' + url+'&maxWidth='+maxWidth
			+ '&inputField=' + input, screen.width * 0.5, screen.height * 0.5,
			input);
}

// file Manager Functions

function fmReturnValue(name) {
	opener.setTargetField(targetField, name);
	window.close();
	return true;
}
function fmDelete(dir, hash, basedir, dirvar, DAT) {
	var resp = confirm('Desea borrar el Archivo?');
	if (resp) {
		window.location = '?modulo=galeriafotos.php&DAT='+DAT+'&dirvar=' + dirvar
				+ '&basedir=' + basedir + '&del=' + hash + '&dir2=' + dir;
	}
}
function fmAddFile(form) {
	var miForm = $(form);
	var newinp = document.createElement('input');
	newinp.setAttribute('type', 'file');
	miForm.append(newinp);

}
// used by fckeditor
function setTargetField(targetField, valor) {
	$('#' + targetField).val(valor);
}

function OpenServerBrowser(url, width, height, input) {
	var iLeft = (screen.width - width) / 2;
	var iTop = (screen.height - height) / 2;

	var sOptions = "toolbar=no,status=no,resizable=yes,dependent=yes";
	sOptions += ",width=" + width;
	sOptions += ",height=" + height;
	sOptions += ",left=" + iLeft;
	sOptions += ",top=" + iTop;

	var oWindow = window.open(url, "BrowseWindow", sOptions);
	oWindow.targetField = input;
}

function SetUrl(url) {
	$('#' + urlobj).val(url);
	oWindow = null;
}

function getURLParameter(url, paramName) {
	var searchString = url, i, val, params = searchString.split("&");

	for (i = 0; i < params.length; i++) {
		val = params[i].split("=");
		if (val[0] == paramName) {
			return unescape(val[1]);
		}
	}
	return null;
}

// VALIDACIONES

function checkNumeric(obj, minval, maxval, comma, period, hyphen) {

	if (chkNumeric(obj, minval, maxval, comma, period, hyphen) == false) {
		obj.select();
		obj.focus();
		return false;
	} else {
		return true;
	}
}

function chkNumeric(obj, minval, maxval, comma, period, hyphen) {
	// only allow 0-9 be entered, plus any values passed
	// (can be in any order, and don't have to be comma, period, or hyphen)
	// if all numbers allow commas, periods, hyphens or whatever,
	// just hard code it here and take out the passed parameters

	var checkOK = "0123456789" + comma + period + hyphen;
	var checkStr = obj;
	var allValid = true;
	var decPoints = 0;
	var allNum = "";

	for (i = 0; i < checkStr.value.length; i++) {
		ch = checkStr.value.charAt(i);
		for (j = 0; j < checkOK.length; j++)
			if (ch == checkOK.charAt(j))
				break;
		if (j == checkOK.length) {
			allValid = false;
			break;
		}
		if (ch != ",")
			allNum += ch;
	}
	if (!allValid) {
		// TODO: REMOVE ID REFERENCE
		// TEST
		var name = $(obj).attr('name');
		alertsay = "Los valores validos son \""
		alertsay = alertsay + checkOK + "\" en el campo \"" + name + "\"."
		alert(alertsay);
		return (false);
	}

	// set the minimum and maximum
	var chkVal = allNum;
	var prsVal = parseInt(allNum);
	if (chkVal != "" && !(prsVal >= minval && prsVal <= maxval)) {
		alertsay = "Por valor ingrese un valor igual "
		alertsay = alertsay + "o mayor que \"" + minval + "\" o igual "
		alertsay = alertsay + "o menos que \"" + maxval + "\" en en campo \""
				+ checkStr.name;
		alert(alertsay);
		return (false);
	}

	return true;
}

/**
 * DHTML date validation script. Courtesy of SmartWebby.com
 * (http://www.smartwebby.com/dhtml/)
 */

function isInteger(s) {
	var i;
	for (i = 0; i < s.length; i++) {
		// Check that current character is number.
		var c = s.charAt(i);
		if (((c < "0") || (c > "9")))
			return false;
	}
	// All characters are numbers.
	return true;
}

function stripCharsInBag(s, bag) {
	var i;
	var returnString = "";
	// Search through string's characters one by one.
	// If character is not in bag, append to returnString.
	for (i = 0; i < s.length; i++) {
		var c = s.charAt(i);
		if (bag.indexOf(c) == -1)
			returnString += c;
	}
	return returnString;
}

function daysInFebruary(year) {
	// February has 29 days in any year evenly divisible by four,
	// EXCEPT for centurial years which are not also divisible by 400.
	return (((year % 4 == 0) && ((!(year % 100 == 0)) || (year % 400 == 0))) ? 29
			: 28);
}
function DaysArray(n) {
	for ( var i = 1; i <= n; i++) {
		this[i] = 31;
		if (i == 4 || i == 6 || i == 9 || i == 11)
			this[i] = 30;
		if (i == 2)
			this[i] = 29
	}
	return this
}

function isDate(obj) {
	// Declaring valid date character, minimum year and maximum year
	var dtCh = "/";

	var minYear = 1900;
	var maxYear = 2100;

	var dtStr = obj.value;
	if (dtStr == '')
		return true;
	var daysInMonth = DaysArray(12);
	var pos1 = dtStr.indexOf(dtCh);
	var pos2 = dtStr.indexOf(dtCh, pos1 + 1);
	var strDay = dtStr.substring(0, pos1);
	var strMonth = dtStr.substring(pos1 + 1, pos2);
	var strYear = dtStr.substring(pos2 + 1);
	var valid = true;
	var error = '';
	strYr = strYear;
	if (strDay.charAt(0) == "0" && strDay.length > 1)
		strDay = strDay.substring(1);
	if (strMonth.charAt(0) == "0" && strMonth.length > 1)
		strMonth = strMonth.substring(1);
	for ( var i = 1; i <= 3; i++) {
		if (strYr.charAt(0) == "0" && strYr.length > 1)
			strYr = strYr.substring(1);
	}
	month = parseInt(strMonth);
	day = parseInt(strDay);
	year = parseInt(strYr);
	if (pos1 == -1 || pos2 == -1) {
		error = "El formato debe ser : dd/mm/aaaa";
		valid = false;
	} else if (strMonth.length < 1 || month < 1 || month > 12) {
		error = "Ingrese un mes valido";
		valid = false;
	} else if (strDay.length < 1 || day < 1 || day > 31
			|| (month == 2 && day > daysInFebruary(year))
			|| day > daysInMonth[month]) {
		error = "Ingrese un dia valido";
		valid = false;
	} else if (strYear.length != 4 || year == 0 || year < minYear
			|| year > maxYear) {
		error = "Ingrese un a&ntilde;o valido de 4 digitos entre " + minYear
				+ " y " + maxYear;
		valid = false;
	} else if (dtStr.indexOf(dtCh, pos2 + 1) != -1
			|| isInteger(stripCharsInBag(dtStr, dtCh)) == false) {
		error = "Fecha Invalida";
		valid = false;
	}

	var Field = $(obj);

	if (valid == false) {
		Field.focus();
		Histrix.alerta(error, obj, 'chan');
		Field.addClass('error').attr("valid", 'false');
		return false;
	} else {
		Field.removeClass('error').attr("valid", 'true');
	}
	return true;
}

/*
 * Para logear en al consola del firebug
 * http://encytemedia.com/blog/articles/2006/05/12/an-in-depth-look-at-the-future-of-javascript-debugging-with-firebug
 */
function loger(string) {
	
	if (console != undefined)
		console.log(string);
}

// Custom js functions
function CPcuitValido(cuit) {
	var vec = new Array(10);
	var esCuit = false;
	var cuit_rearmado = "";
	errors = ''
	for (i = 0; i < cuit.length; i++) {
		caracter = cuit.charAt(i);
		if (caracter.charCodeAt(0) >= 48 && caracter.charCodeAt(0) <= 57) {
			cuit_rearmado += caracter;
		}
	}
	cuit = cuit_rearmado;
	if (cuit.length != 11) { // si to estan todos los digitos
		esCuit = false;
		errors = 'Cuit <11 ';
		return "CUIT Menor a 11 Caracteres";
	} else {
		x = i = dv = 0;
		// Multiplico los dígitos.
		vec[0] = cuit.charAt(0) * 5;
		vec[1] = cuit.charAt(1) * 4;
		vec[2] = cuit.charAt(2) * 3;
		vec[3] = cuit.charAt(3) * 2;
		vec[4] = cuit.charAt(4) * 7;
		vec[5] = cuit.charAt(5) * 6;
		vec[6] = cuit.charAt(6) * 5;
		vec[7] = cuit.charAt(7) * 4;
		vec[8] = cuit.charAt(8) * 3;
		vec[9] = cuit.charAt(9) * 2;

		// Suma cada uno de los resultado.
		for (i = 0; i <= 9; i++) {
			x += vec[i];
		}
		dv = (11 - (x % 11)) % 11;
		if (dv == cuit.charAt(10)) {
			esCuit = true;
		}
	}
	if (!esCuit) {
		return "CUIT Invalido";
	}
	return true;
}

/** jQuery next field plugin */
$.fn.focusNextInputField = function () {

	return this
			.each(function () {
				var fields = $(this)
						.parents('form:eq(0),body')
						.find(
								'button:visible:enabled[tabindex],input:visible:enabled,textarea:enabled,select:visible:enabled');
				if (fields.length == 0) {
					fields = $(this)
							.parents('.contenido')
							.find(
									'button:visible:enabled[tabindex],input:visible:enabled,textarea:enabled,select:visible:enabled');
				}

				var index = fields.index(this);

				if (index > -1 && (index + 1) < fields.length) {
					fields.eq(index + 1).focus().select();
				}
				return false;
			});
};

// / MAX LENGTH
jQuery.fn.maxLength = function () {
	this
			.each(function () {
				max = $(this).attr('maxlength');
				if (max == 0)
					return;
				// Get the type of the matched element
				var type = this.tagName.toLowerCase();
				// If the type property exists, save it in lower case
				var inputType = this.type ? this.type.toLowerCase() : null;
				// Check if is a input type=text OR type=password
				if (type == "input" && inputType == "text"
						|| inputType == "password") {
					// Apply the standard maxLength
					this.maxLength = max;
				}
				// Check if the element is a textarea
				else if (type == "textarea") {
					// Add the key press event
					this.onkeypress = function (e) {
						// refresh max
						max = $(this).attr('maxlength');
						if (max == 0)
							return;
						// Get the event object (for IE)
						var ob = e || event;
						// Get the code of key pressed
						var keyCode = ob.keyCode;
						// Check if it has a selected text
						var hasSelection = document.selection ? document.selection
								.createRange().text.length > 0
								: this.selectionStart != this.selectionEnd;
						// return false if can't write more
						var response = !(this.value.length >= max
								&& (keyCode > 50 || keyCode == 32
										|| keyCode == 0 || keyCode == 13)
								&& !ob.ctrlKey && !ob.altKey && !hasSelection);
						if (response == false) {
							Histrix.alerta('max size' + max, $(this));
						}

						return response;
					};
					// Add the key up event
					this.onkeyup = function () {
						// refresh max size
						var $this = $(this);
						max = $this.attr('maxlength');
						if (max == 0)
							return;
						// If the keypress fail and allow write more text that
						// required, this event will remove it
						if (this.value.length > max) {
							this.value = this.value.substring(0, max);
						}
						// loger($('#_FM_' + $(this).attr('name')));
						if (!$('#_FM_' + $this.attr('name'))[0]) {

							$this.after('<div class="msg" id="_FM_'
									+ $this.attr('name') + '"></div>');
							setTimeout( 
								function () {
									$("#_FM_" + $this.attr('name')).remove();
								}, 5000);
						}

						$('#_FM_' + $this.attr('name')).html(
								(this.value.length) + '/' + max);
					};
				}
			});
};

function parseURL(url) {
	var a = document.createElement('a');
	a.href = url;
	return {
		source : url,
		protocol : a.protocol.replace(':', ''),
		host : a.hostname,
		port : a.port,
		query : a.search,
		params : (function () {
			var ret = {}, seg = a.search.replace(/^\?/, '').split('&'), len = seg.length, i = 0, s;
			for (; i < len; i++) {
				if (!seg[i]) {
					continue;
				}
				s = seg[i].split('=');
				ret[s[0]] = s[1];
			}
			return ret;
		})(),
		file : (a.pathname.match(/\/([^\/?#]+)$/i) || [ , '' ])[1],
		hash : a.hash.replace('#', ''),
		path : a.pathname.replace(/^([^\/])/, '/$1'),
		relative : (a.href.match(/tps?:\/\/[^\/]+(.+)/) || [ , '' ])[1],
		segments : a.pathname.replace(/^\//, '').split('/')
	};
} 
// 5413 (14/09/2010) 
// 5723 (27/12/2010) 
// 5661 (12/01/2011) 
// 5816 (22/01/2011) 
// 5893 (13/02/2011) 
// 6376 (14/08/2011) 
// 6500 (02/10/2011)
// 6893 (10/01/2012)
// 7146 (18/04/2012)
// 7361 (09/11/2012)
