// the assumed x ratio component of the video
var ASSUMED_X_VIDEO_RATIO = 5;
// the assumed y ratio component of the video
var ASSUMED_Y_VIDEO_RATIO = 3;
// the screen ratio from the last video style rendering
var LAST_RATIO;
// a threshold for how much the screen ratio must change in order
// to trigger a resize event
var RATIO_CHANGE_THRESHOLD = 0.1;
// maps a video element with its location in the video container
var videoIndices = {};

/**
 * Initialize the UI with a good display
 */
window.onload = function() {
    var screensize = getScreenSize();
    LAST_RATIO = screensize[0] / screensize[1];
    updateVideoLocations();
};

/**
 * Only triggers a video re-render if the screen ratio
 * has changed sufficiently
 */
window.onresize = function() {
    // TODO: make a better metric for this
    var screensize = getScreenSize();
    if (Math.abs(LAST_RATIO - (screensize[0] / screensize[1])) > RATIO_CHANGE_THRESHOLD * LAST_RATIO) {
        console.log("triggering video resize event");
        updateVideoLocations();
    }
};

function updateVideoLocations() {
    var boxes = document.getElementById("remoteVideos").children;
    var screensize = getScreenSize();

    LAST_RATIO = screensize[0] / screensize[1];

    var styles = getGridDivStyles(
        ASSUMED_X_VIDEO_RATIO,
        ASSUMED_Y_VIDEO_RATIO,
        boxes.length,
        screensize[0],
        screensize[1],
        videoIndices[videoIndices.primary]
    );
    var one_style;
    var box;

    // re-assign styles to existing grids, creating new ones as necessary
    for (var i = 0; i < styles.length; i++) {
        box = boxes[i];
        one_style = styles[i];
        for (var attribute in one_style)
            box.style[attribute] = one_style[attribute];
        // assign a position in the video map
        videoIndices[box.id] = i;
    }


}

function getGridDivStyles(aspect_x, aspect_y, n_boxes, screen_px_x, screen_px_y, active_box) {
    if (!active_box || active_box < 0 || active_box >= n_boxes || n_boxes == 1)
        return getGridDivStylesForRegion(aspect_x, aspect_y, n_boxes, 100, 100, 0, 0, screen_px_x, screen_px_y);

    // split into regions and assemble the screens back together. the screen
    //	regions are described below:
    //
    //	+-------------------------------+
    //	|	4+ boxes: r1 == active box  |
    //	|	________________			|
    //	|	|	   	 |		| 			|
    //	|	|___r1___|		|			|
    //	|	|	     |		|			|
    //	|	|___r2___|__r3__|			|
    //	|								|
    //	|- - - - - - - - - - - - - - - -|
    //	|	2-4 boxes: r1 == active box |
    //	|	________________			|
    //	|	|		    |	| 			|
    //	|	|			|	|			|
    //	|	|	  		|	|			|
    //	|	|____r1_____|_r2|			|
    //	|								|
    //	+-------------------------------+
    //

    var active_weight, total_weight;

    // assign weights dependign on cases
    if (n_boxes == 2) {
        active_weight = 8;
        total_weight = 10;
    } else if (n_boxes == 3) {
        active_weight = 8;
        total_weight = 11;
    } else {
        active_weight = n_boxes + 1.5;
        total_weight = n_boxes * 2;
    }

    var active_scale = active_weight / total_weight;
    var r1, r2, r3;
    // r#: [n_box_in_region, %_region_width, %_region_height, %_top_left_x, %_top_left_y, px_region_width, px_region_height]
    if (n_boxes == 2 || n_boxes == 3 || n_boxes == 4) {
        r1 = [
            1,
            100 * active_scale,
            100,
            0,
            0,
            screen_px_x * active_scale,
            screen_px_y
        ];

        r2 = [
            n_boxes - 1,
            100 * (1 - active_scale),
            100,
            100 * active_scale,
            0,
            screen_px_x * (1 - active_scale),
            screen_px_y
        ];
    } else {
        r1 = [
            1,
            100 * active_scale,
            100 * active_scale,
            0,
            0,
            screen_px_x * active_scale,
            screen_px_y * active_scale
        ];

        r2 = [
            Math.ceil((active_scale * (1 - active_scale) * (n_boxes - 1))) + 1,
            100 * active_scale,
            100 * (1 - active_scale),
            0,
            100 * active_scale,
            screen_px_x * active_scale,
            screen_px_y * (1 - active_scale),
        ];

        r3 = [
            n_boxes - r1[0] - r2[0],
            100 * (1 - active_scale),
            100,
            100 * active_scale,
            0,
            screen_px_x * (1 - active_scale),
            screen_px_y
        ];
    }

    var r1_styles, r2_styles, r3_styles, all_styles;
    r1_styles = getGridDivStylesForRegion(aspect_x, aspect_y, r1[0], r1[1], r1[2], r1[3], r1[4], r1[5], r1[6]);
    r2_styles = getGridDivStylesForRegion(aspect_x, aspect_y, r2[0], r2[1], r2[2], r2[3], r2[4], r2[5], r2[6]);
    all_styles = r1_styles.concat(r2_styles);

    if (n_boxes > 4) {
        r3_styles = getGridDivStylesForRegion(aspect_x, aspect_y, r3[0], r3[1], r3[2], r3[3], r3[4], r3[5], r3[6]);
        all_styles = all_styles.concat(r3_styles);
    }

    // move the active box into the first style slot
    var tmp = all_styles[0];
    all_styles[0] = all_styles[active_box];
    all_styles[active_box] = tmp;

    return all_styles;
}

// return grid styles for a screen region.
//	aspect_x
//		x aspect ratio for video box
//	aspect_y
//		y aspect ratio for video box
//	n_boxes
//		number of video boxes for this region
//	screen_x
//		the % width of the screen region
//	screen_y
//		the % height of the screen region
//	top_left_x
//		the top left width % point of the screen region
//	top_left_y
//		the top left height % point of the screen region
//	screen_px_x
//		the total region width, in pixels
//	screen_px_y
//		the total region height, in pixels
function getGridDivStylesForRegion(aspect_x, aspect_y, n_boxes, screen_x, screen_y, top_left_x, top_left_y, screen_px_x, screen_px_y) {
    var dims = getGridDims(aspect_x, aspect_y, n_boxes, screen_px_x, screen_px_y);
    var w = dims[0];
    var h = dims[1];
    var n_grid_points = w * h;
    var fits = w * h >= n_boxes;
    var styles = [];

    console.log(n_boxes + " box(es) (cols x rows) == (" + w + " x " + h + ") (" + fits + ")");
    if (!fits)
        alert("Boxes wont fit!");

    // calculate which rows should have a different effective size,
    // due to empty grid elements.
    var modified_size_rows = {};
    var n_modified_rows = (n_grid_points - n_boxes);
    var n_top_modified_rows;
    var n_btm_modified_rows;

    // if more rows are modified, then put unmodified rows between
    // the modified rows because it looks nicer. otherwise, put the
    // modified rows between the unmodified rows
    if (n_modified_rows > h - n_modified_rows)
        n_top_modified_rows = Math.floor((h - 1) / 2);
    else
        n_top_modified_rows = Math.ceil(n_modified_rows / 2);

    n_btm_modified_rows = n_modified_rows - n_top_modified_rows;
    var i;
    for (i = 0; i < n_top_modified_rows; i++){
        modified_size_rows[i] = true;
    }
    for (i = 0; i < n_btm_modified_rows; i++){
        modified_size_rows[h - 1 - i] = true;
    }
    for (var r = 0; r < h; r++) {
        for (var c = 0; c < w; c++) {
            if (styles.length < n_boxes) {
                var style = {};

                var effective_h = h;
                var effective_w = w;

                if (r in modified_size_rows)
                    effective_w--;

                var width = screen_x / effective_w;
                var height = screen_y / effective_h;
                var left = c * width;
                var top = r * height;

                // a change in effective height or width would push this grid item
                // off the screen, so we can move it to the next row. Add a pertubation
                // to avoid issues with float arithmetic
                if (left + width > screen_x + 0.0001 || top + height > screen_y + 0.0001)
                    continue;

                style.width = width + '%';
                style.height = height + '%';
                style.top = top + top_left_y + '%';
                style.left = left + top_left_x + '%';
                style.position = 'absolute';
                styles.push(style);
            }
        }
    }
    return styles;
}


// returns the aproximate best grid dimensions for a given screen
//	box_x_ratio
//		x aspect ratio for video box
//	box_y_ratio
//		y aspect ratio for video box
//	n_box
//		number of video boxes for this region
//	screen_px_x
//		the total region width, in pixels
//	screen_px_y
//		the total region height, in pixels
function getGridDims(box_x_ratio, box_y_ratio, n_box, screen_px_x, screen_px_y) {

    // gives a good aproximation for grid dimensions
    var r = (box_x_ratio * screen_px_y) / (box_y_ratio * screen_px_x);
    var n_cols = Math.ceil(Math.sqrt(n_box / r));
    var n_rows = Math.ceil(n_box / n_cols);

    // shrinks grid dimensions as much as possible
    var made_change = true;
    var shrink_rows = screen_px_x > screen_px_y ? true : false;
    while (n_cols * n_rows > n_box && made_change) {
        made_change = false;
        if ((n_rows - 1) * n_cols >= n_box) {
            n_rows--;
            made_change = true;
        }
        if (n_rows * (n_cols - 1) >= n_box) {
            n_cols--;
            made_change = true;
        }
    }

    return [
        (isFinite(n_cols) ? n_cols : 1.0) || 1,
        (isFinite(n_rows) ? n_rows : 1.0) || 1
    ];
}

function getScreenSize() {
    var w = $('body').innerWidth();
    var h = $('body').innerHeight();
    return [w, h];
}
