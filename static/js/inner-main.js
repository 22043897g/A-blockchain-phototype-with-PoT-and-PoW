(function ($) {
  "use strict";

  // data - background
  $("[data-background]").each(function () {
    $(this).css("background-image", "url(" + $(this).attr("data-background") + ")")
  })

  // active
  $('.pt-wrp').on('mouseenter', function () {
    $(this).addClass('active').parent().siblings().find('.pt-wrp').removeClass('active');
  })

  // meanmenu
  $('#mobile-menu').meanmenu({
    meanMenuContainer: '.mobile-menu',
    meanScreenWidth: "992",
    onePage: true,
  });

  $(window).on('scroll', function () {
    var scroll = $(window).scrollTop();
    if (scroll < 245) {
      $("#header-sticky").removeClass("sticky-bar");
    } else {
      $("#header-sticky").addClass("sticky-bar");
    }
  });


// One Page Nav
var top_offset = $('.header-area').height() - 0;
 	$('.main-menu nav ul').onePageNav({
   	currentClass: 'active',
   	scrollOffset: top_offset,
});

  /*---------------------
        Circular Bars - Knob
     --------------------- */

  if (typeof ($.fn.knob) != 'undefined') {
    $('.knob').each(function () {
      var $this = $(this),
        knobVal = $this.attr('data-rel');

      $this.knob({
        'draw': function () {
          $(this.i).val(this.cv + '%')
        }
      });

      $this.appear(function () {
        $({
          value: 0
        }).animate({
          value: knobVal
        }, {
          duration: 2000,
          easing: 'swing',
          step: function () {
            $this.val(Math.ceil(this.value)).trigger('change');
          }
        });
      }, {
        accX: 0,
        accY: -150
      });
    });
  }


  $('.popup-img').magnificPopup({
    type: 'image'
  });

  $('.popup-video').magnificPopup({
    type: 'iframe'
  });

  $('.testimonail-active').slick({
    dots: false,
    arrows: false,
    infinite: true,
    speed: 300,
    slidesToShow: 1,
    slidesToScroll: 1,
    responsive: [{
      breakpoint: 1024,
      settings: {
        slidesToShow: 1,
        slidesToScroll: 1,
        infinite: true,
        dots: false,
      }
    },
    {
      breakpoint: 992,
      settings: {
        slidesToShow: 1,
        slidesToScroll: 2
      }
    },
    {
      breakpoint: 600,
      settings: {
        slidesToShow: 1,
        slidesToScroll: 2
      }
    },
    {
      breakpoint: 480,
      settings: {
        slidesToShow: 1,
        slidesToScroll: 1
      }
    }
    ]
  });

  $('.brand-active').slick({
    dots: false,
    arrows: false,
    infinite: true,
    speed: 300,
    slidesToShow: 6,
    slidesToScroll: 1,
    responsive: [{
      breakpoint: 1200,
      settings: {
        slidesToShow: 4,
        slidesToScroll: 1,
        infinite: true,
        dots: false,
      }
    },
    {
      breakpoint: 992,
      settings: {
        slidesToShow: 4,
        slidesToScroll: 2
      }
    },
    {
      breakpoint: 600,
      settings: {
        slidesToShow: 2,
        slidesToScroll: 2
      }
    },
    {
      breakpoint: 550,
      settings: {
        slidesToShow: 1,
        slidesToScroll: 1
      }
    }
    ]
  });


})(jQuery);