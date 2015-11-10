/*!
 * Modals v8.1.0: Simple modal dialogue pop-up windows
 * (c) 2015 Chris Ferdinandi
 * MIT License
 * http://github.com/cferdinandi/modals
 */

(function (root, factory) {
	if ( typeof define === 'function' && define.amd ) {
		define([], factory(root));
	} else if ( typeof exports === 'object' ) {
		module.exports = factory(root);
	} else {
		root.modals = factory(root);
	}
})(typeof global !== 'undefined' ? global : this.window || this.global, function (root) {

	'use strict';

	//
	// Variables
	//

	var publicApi = {}; // Object for public APIs
	var supports = 'querySelector' in document && 'addEventListener' in root && 'classList' in document.createElement('_'); // Feature test
	var state = 'closed';
	var settings;

	// Default settings
	var defaults = {
		selectorToggle: '[data-modal]',
		selectorWindow: '[data-modal-window]',
		selectorClose: '[data-modal-close]',
		modalActiveClass: 'active',
		modalBGClass: 'modal-bg',
		backspaceClose: true,
		callbackOpen: function () {},
		callbackClose: function () {}
	};


	//
	// Methods
	//

	/**
	 * A simple forEach() implementation for Arrays, Objects and NodeLists.
	 * @private
	 * @author Todd Motto
	 * @link   https://github.com/toddmotto/foreach
	 * @param {Array|Object|NodeList} collection Collection of items to iterate
	 * @param {Function}              callback   Callback function for each iteration
	 * @param {Array|Object|NodeList} scope      Object/NodeList/Array that forEach is iterating over (aka `this`)
	 */
	var forEach = function ( collection, callback, scope ) {
		if ( Object.prototype.toString.call( collection ) === '[object Object]' ) {
			for ( var prop in collection ) {
				if ( Object.prototype.hasOwnProperty.call( collection, prop ) ) {
					callback.call( scope, collection[prop], prop, collection );
				}
			}
		} else {
			for ( var i = 0, len = collection.length; i < len; i++ ) {
				callback.call( scope, collection[i], i, collection );
			}
		}
	};

	/**
	 * Merge two or more objects. Returns a new object.
	 * @private
	 * @param {Boolean}  deep     If true, do a deep (or recursive) merge [optional]
	 * @param {Object}   objects  The objects to merge together
	 * @returns {Object}          Merged values of defaults and options
	 */
	var extend = function () {

		// Variables
		var extended = {};
		var deep = false;
		var i = 0;
		var length = arguments.length;

		// Check if a deep merge
		if ( Object.prototype.toString.call( arguments[0] ) === '[object Boolean]' ) {
			deep = arguments[0];
			i++;
		}

		// Merge the object into the extended object
		var merge = function (obj) {
			for ( var prop in obj ) {
				if ( Object.prototype.hasOwnProperty.call( obj, prop ) ) {
					// If deep merge and property is an object, merge properties
					if ( deep && Object.prototype.toString.call(obj[prop]) === '[object Object]' ) {
						extended[prop] = extend( true, extended[prop], obj[prop] );
					} else {
						extended[prop] = obj[prop];
					}
				}
			}
		};

		// Loop through each object and conduct a merge
		for ( ; i < length; i++ ) {
			var obj = arguments[i];
			merge(obj);
		}

		return extended;

	};

	/**
	 * Get the closest matching element up the DOM tree.
	 * @private
	 * @param  {Element} elem     Starting element
	 * @param  {String}  selector Selector to match against (class, ID, data attribute, or tag)
	 * @return {Boolean|Element}  Returns null if not match found
	 */
	var getClosest = function ( elem, selector ) {

		// Variables
		var firstChar = selector.charAt(0);
		var attribute, value;

		// If selector is a data attribute, split attribute from value
		if ( firstChar === '[' ) {
			selector = selector.substr(1, selector.length - 2);
			attribute = selector.split( '=' );

			if ( attribute.length > 1 ) {
				value = true;
				attribute[1] = attribute[1].replace( /"/g, '' ).replace( /'/g, '' );
			}
		}

		// Get closest match
		for ( ; elem && elem !== document; elem = elem.parentNode ) {

			// If selector is a class
			if ( firstChar === '.' ) {
				if ( elem.classList.contains( selector.substr(1) ) ) {
					return elem;
				}
			}

			// If selector is an ID
			if ( firstChar === '#' ) {
				if ( elem.id === selector.substr(1) ) {
					return elem;
				}
			}

			// If selector is a data attribute
			if ( firstChar === '[' ) {
				if ( elem.hasAttribute( attribute[0] ) ) {
					if ( value ) {
						if ( elem.getAttribute( attribute[0] ) === attribute[1] ) {
							return elem;
						}
					} else {
						return elem;
					}
				}
			}

			// If selector is a tag
			if ( elem.tagName.toLowerCase() === selector ) {
				return elem;
			}

		}

		return null;

	};

	/**
	 * Stop YouTube, Vimeo, and HTML5 videos from playing when leaving the slide
	 * @private
	 * @param  {Element} content The content container the video is in
	 * @param  {String} activeClass The class asigned to expanded content areas
	 */
	var stopVideos = function ( content, activeClass ) {
		if ( !content.classList.contains( activeClass ) ) {
			var iframe = content.querySelector( 'iframe');
			var video = content.querySelector( 'video' );
			if ( iframe ) {
				var iframeSrc = iframe.src;
				iframe.src = iframeSrc;
			}
			if ( video ) {
				video.pause();
			}
		}
	};

	/**
	 * Open the target modal window
	 * @public
	 * @param  {Element} toggle The element that toggled the open modal event
	 * @param  {String} modalID ID of the modal to open
	 * @param  {Object} options
	 * @param  {Event} event
	 */
	publicApi.openModal = function (toggle, modalID, options) {

		// Define the modal
		var settings = extend( settings || defaults, options || {} );  // Merge user options with defaults
		var modal = document.querySelector(modalID);

		// Define the modal background
		var modalBg = document.createElement('div');
		modalBg.setAttribute('data-modal-bg', null);
		modalBg.classList.add( settings.modalBGClass );

		// Activate the modal
		modal.classList.add( settings.modalActiveClass );
		document.body.appendChild(modalBg);
		state = 'open';

		settings.callbackOpen( toggle, modalID ); // Run callbacks after opening a modal

	};

	/**
	 * Close all modal windows
	 * @public
	 * @param  {Object} options
	 * @param  {Event} event
	 */
	publicApi.closeModals = function (toggle, options) {

		// Selectors and variables
		var settings = extend( defaults, options || {} ); // Merge user options with defaults
		var openModals = document.querySelectorAll( settings.selectorWindow + '.' + settings.modalActiveClass ); // Get open modals
		var modalsBg = document.querySelectorAll( '[data-modal-bg]' ); // Get modal background element

		if ( openModals.length > 0 || modalsBg.length > 0 ) {

			// Close all modals
			forEach(openModals, function (modal) {
				if ( modal.classList.contains( settings.modalActiveClass ) ) {
					stopVideos(modal); // If active, stop video from playing
					modal.classList.remove( settings.modalActiveClass );
				}
			});

			// Remove all modal backgrounds
			forEach(modalsBg, function (bg) {
				document.body.removeChild(bg);
			});

			// Set state to closed
			state = 'closed';

			settings.callbackClose(); // Run callbacks after closing a modal

		}

	};

	/**
	 * Handle toggle click events
	 * @private
	 */
	var eventHandler = function (event) {
		var toggle = event.target;
		var open = getClosest(toggle, settings.selectorToggle);
		var close = getClosest(toggle, settings.selectorClose);
		var modal = getClosest(toggle, settings.selectorWindow);
		var key = event.keyCode;

		if ( key && state === 'open' ) {
			if ( key === 27 || ( settings.backspaceClose && ( key === 8 || key === 46 ) ) ) {
				publicApi.closeModals(null, settings);
			}
		} else if ( toggle ) {
			if ( modal && !close ) {
				return;
			} else if ( open ) {
				event.preventDefault();
				publicApi.openModal( open, open.getAttribute('data-modal'), settings );
			} else if ( state === 'open' ) {
				event.preventDefault();
				publicApi.closeModals(toggle, settings);
			}
		}
	};

	/**
	 * Destroy the current initialization.
	 * @public
	 */
	publicApi.destroy = function () {
		if ( !settings ) return;
		document.removeEventListener('click', eventHandler, false);
		document.removeEventListener('touchstart', eventHandler, false);
		document.removeEventListener('keydown', eventHandler, false);
		settings = null;
	};

	/**
	 * Initialize Modals
	 * @public
	 * @param {Object} options User settings
	 */
	publicApi.init = function ( options ) {

		// feature test
		if ( !supports ) return;

		// Destroy any existing initializations
		publicApi.destroy();

		// Merge user options with defaults
		settings = extend( defaults, options || {} );

		// Listen for events
		document.addEventListener('click', eventHandler, false);
		document.addEventListener('touchstart', eventHandler, false);
		document.addEventListener('keydown', eventHandler, false);

	};


	//
	// Public APIs
	//

	return publicApi;

});