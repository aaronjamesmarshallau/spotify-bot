var Spotify = (function () {
    var _ = {
        consts: {
            SEARCH_ENDPOINT: "/search?q=",
            ALBUMS_ENDPOINT: "/albums?id=",
            TRACKS_ENDPOINT: "/tracks?id="
        },
        currentStatus: null,
        currentQueue: [],
        hasQueueChanged: function (newQueue) {
            if (!newQueue) {
                return false;
            }

            if (newQueue.length != _.currentQueue.length) {
                return true;
            }

            for (var i = 0; i < newQueue.length; i++) {
                if (newQueue[i].trackId !== _.currentQueue[i].trackId) {
                    return true;
                }
            }

            return false;
        },
        loginPage: function () {
            return $("<div></div>", {
                class: "login-page",
                html: [
                    $("<div></div>", {
                        class: "login-page-inner",
                        html: [
                            $("<div></div>", {
                                class: "login-form-container",
                                html: [
                                    $("<form></form>", {
                                        class: "login-form",
                                        html: [
                                            $("<input/>", {
                                                name: "uid",
                                                type: "password",
                                                placeholder: "Unique ID"
                                            }),
                                            $("<button></button>", {
                                                class: "login-button",
                                                text: "Login",
                                                type: "button",
                                                on: {
                                                    click: function () {
                                                        $.ajax({
                                                            url: "/auth",
                                                            data: {
                                                                authId: $("input[name=uid]").val()
                                                            },
                                                            success: function () {
                                                                $(".login-page").remove();
                                                                spotify.init();
                                                                spotify.start();
                                                            },
                                                            error: function () {
                                                                $(".error").text("Invalid UID");
                                                            }
                                                        })
                                                    }
                                                }
                                            }),
                                            $("<span></span>", {
                                                class: "error"
                                            })
                                        ]
                                    })
                                ]
                            })
                        ]
                    })
                ]
            });
        },
        onTrackChanged: function () {
            window.localStorage.setItem("hasVoted", false);

			_.currentQueue.shift();
			_.onQueueChange(_.currentQueue);
			_.updateTrackUi();
        },
        onQueueChange: function (newQueue) {
            _.currentQueue = newQueue;
			$(".queue-list").empty();

			for (var i = 0; i < _.currentQueue.length; i++)
			(function (index) {
				var track = _.currentQueue[index];
				var dom = $("<li></li>", {
					"data-track-id": track.trackId,
					class: "queue-item",
					html: [
                        $("<div></div>", {
                            class: "queue-container",
                            html: [
                                $("<div></div>", {
                                    class: "queue-album-art",
                                    html: [
                                        $("<img/>", {
                                            src: track.albumArt,
                                            alt: track.albumName
                                        })
                                    ]
                                }),
                                $("<div></div>", {
                                    class: "queue-track-info",
                                    html: [
                                        $("<span></span>", {
                                            class: "queue-track-name",
                                            text: track.trackName
                                        }),
                                        $("<span></span>", {
                                            class: "queue-track-artist",
                                            text: track.artistName
                                        }),
                                        $("<span></span>", {
                                            class: "queue-track-album",
                                            text: track.albumName
                                        }),
                                    ]
                                })
                            ]
                        })
					]
				});

				$(".queue-list").append(dom);
			})(i);
        },
        updatePlayingUi: function () {
            var playing = _.currentStatus.playing;
			var currentPlayPosition = (_.currentStatus.playing_position / _.currentStatus.track.length) * 100;

			$(".track-duration-track").css("width", currentPlayPosition + "%");

			$(".upvote-count").text(_.currentStatus.currentUpvotes);
			$(".downvote-count").text(_.currentStatus.currentDownvotes);

			if (playing) {
				$(".spotify-button.playpause img").attr("src", "./images/pause.svg");
			} else {
				$(".spotify-button.playpause img").attr("src", "./images/play.svg");
			}
        },
        updateTrackUi: function () {
            var albumId = _.currentStatus.track.album_resource.uri.replace("spotify:album:", "");
			var currentTrackTitle = _.currentStatus.track.track_resource.name;
			var currentArtistTitle = _.currentStatus.track.artist_resource.name;
			var currentAlbumTitle = _.currentStatus.track.album_resource.name;
			var nowPlayingAreaHeight = $(".playing-panel-inner").height() - 20;

			spotify.getAlbumArt(albumId, function (imageUri) {
				$(".album-artwork").attr("src", imageUri);
			});

			$(".now-playing-container").height(nowPlayingAreaHeight);
			$(".now-playing-container").width(nowPlayingAreaHeight);

			$(".song-title").text(currentTrackTitle);
			$(".artist-title").text(currentArtistTitle);
			$(".album-title").text(currentAlbumTitle);
        }
    };

    var spotify = {
        getStatus: function (callback) {
            $.ajax({
				url: "/status",
				method: "GET",
				success: function (data) {
					if (_.currentStatus != null) {
						var currentTrack = _.currentStatus.track.track_resource.uri.replace("spotify:album:", "");
						var newTrack = data.track.track_resource.uri.replace("spotify:album:", "");

						_.currentStatus = data;

						if (currentTrack !== newTrack) {
							_.onTrackChanged();
						}
					} else {
						_.currentStatus = data;
						_.onTrackChanged();
					}

					if (typeof callback === "function") callback();
				}
			});
        },
        getAlbumArt: function (albumId, callback) {
			$.ajax({
				url: _.consts.ALBUMS_ENDPOINT + albumId,
				success: function (response) {
					var imageUri = response.images[0].url;

					callback(imageUri);
				}
			});
		},
        getTrackInfo: function (trackId, callback) {
            $.ajax({
                url: _.consts.TRACKS_ENDPOINT + trackId,
                method: "GET",
                success: function (response) {
                    var track = {
                        albumName: response.album.name,
                        albumImage: response.album.images[1].url,
                        artistName: response.artists[0].name,
                        trackName: response.name
                    };

                    callback(track);
                }
            });
        },
        refreshQueue: function () {
            $.ajax({
				url: '/queue',
				method: 'GET',
				success: function (response) {
					if (_.hasQueueChanged(response)) {
						_.onQueueChange(response);
					}
				}
			});
        },
        search: function (text) {
            if (text.length == 0) {
				$(".search-results").remove();
				return;
			}

			$.ajax({
				url: _.consts.SEARCH_ENDPOINT + text,
				error: function () {
					$(".search-results").remove();
				},
				success: function (data) {
					var tracks = data.tracks.items;
					var results = [];
					$(".search-results").remove();

					tracks.sort(function (a, b) {
						if (a.popularity < b.popularity) {
							return 1;
						}

						if (a.popularity > b.popularity) {
							return -1;
						}

						return 0;
					});

					for (var i = 0; i < tracks.length; i++) {
						(function (index) {
							var el = tracks[index];
							var trackId = el.id;
                            var albumName = el.album.name;
							var albumArtSmall = el.album.images[2].url;
                            var albumArt = el.album.images[0].url;
							var trackName = el.name;
							var trackArtist = el.artists[0].name;
							var queueUrl = "/queue";

							var el = $("<li></li>", {
								class: "search-results-item",
								html: [
									$("<div></div>", {
										class: "results-item-image-container",
										html: [
											$("<img/>", {
												src: albumArtSmall
											})
										]
									}),
									$("<div></div>", {
										class: "results-item-info",
										html: [
											$("<div></div>", {
												class: "item-info-title",
												text: trackName
											}),
											$("<div></div>", {
												class: "item-info-artist",
												text: trackArtist
											})
										]
									}),
									$("<div></div>", {
										class: "results-item-queue",
										html: [
											$("<button></button>", {
												class: "queue-button",
												html: [
													$("<img/>", {
														src: "./images/add.svg"
													})
												],
												on: {
													click: function () {
														_.currentQueue.push(trackId);
														_.onQueueChange(_.currentQueue);

                                                        var trackInfo = {
                                                            trackId: trackId,
                                                            trackName: trackName,
                                                            artistName: trackArtist,
                                                            albumArt: albumArt,
                                                            albumName: albumName
                                                        };

														$.ajax({
															url: queueUrl,
                                                            method: "POST",
                                                            dataType: "json",
                                                            contentType: "application/json; charset=UTF-8",
                                                            data: JSON.stringify(trackInfo),
															success: function (data) {
																var el = $("<div></div>", {
																	class: "queue-added-flyout",
																	style: "opacity: 0",
																	html: [
																		$("<div></div>", {
																			class: "flyout-inner",
																			html: [
																				$("<img/>", {
																					class: "flyout-image",
																					src: "./images/queued.svg"
																				}),
																				$("<span></span>", {
																					class: "flyout-text",
																					text: "Added"
																				})
																			]
																		})
																	]
																});

																$("body").append(el);

																$(el).animate({
																	opacity: 1
																}, 500, function () {
																	setTimeout(function () {
																		$(el).fadeOut(200);
																	}, 1000);
																})
															}
														});

														$(".spotify-search input").val("");
														$(".search-results").remove();
													}
												}
											})
										]
									})
								]
							});

							results.push(el);
						})(i);
					}

					var searchResults = $("<ul></ul>", {
                        class: 'search-results'
                    }).append(results);

                    $(".search-pane").append(searchResults);
				}
			});
        },
        showLoginPage: function () {
            $("body").append(_.loginPage());
        },
        assertAuthStatus: function (authCallback, unauthCallback) {
            var me = this;

            $.ajax({
        		url: "/auth",
        		method: "GET",
        		success: function (response, textStatus) {
        			if (!response.auth) {
        				unauthCallback();
        			} else {
        				authCallback();
        			}
        		}
        	});
        },
        init: function () {
            $(".playpause").on('click', function () {
    			var endpoint = _.currentStatus != null
    				? _.currentStatus.playing
    					? "/pause"
    					: "/unpause"
    				: "/unpause";
    			$.ajax({
    				url: endpoint
    			});
    		});

    		$(".queue-button").on('click', function () {
    			$(".queue-panel").toggleClass("shown");
    		});

    		$(".downvote").on("click", function () {
    			if (window.localStorage.getItem("hasVoted") == "true") return;

    			var count = Number($(".downvote-count").text());

    			count++;

    			window.localStorage.setItem("hasVoted", true);

    			$(".downvote-count").text(count);

    			$.ajax({
    				url: "/downvote"
    			});
    		});

            $(window).on("resize", function () {
                _.updatePlayingUi();
                _.updateTrackUi();
            });

    		$(".upvote").on("click", function () {
    			if (window.localStorage.getItem("hasVoted") == "true") return;

    			var count = Number($(".upvote-count").text());

    			count++;

    			$(".upvote-count").text(count);

    			window.localStorage.setItem("hasVoted", true);

    			$.ajax({
    				url: "/upvote"
    			});
    		});


    		var searchInput = $(".spotify-search input");
    		var searchImg = $(".spotify-search img");
            var timeout;
            var mouseentered = false;

            searchInput.on({
                keyup: function (evt) {
                    if (evt.keyCode == 27) {
                        $(this).val("");
                        $('.search-results').remove();
                        return;
                    }

        			var text = $(this).val()

                    if (timeout != null) {
                        clearTimeout(timeout);
                    }

        			timeout = setTimeout(function () {
        				spotify.search(text);
        			}, 200);
        		},
                mouseenter: function () {
                    mouseentered = true;
                    searchInput.css({'opacity': 1});
                    searchImg.css({'opacity': 0});
                },
                mouseleave: function () {
                    mouseentered = false;
                    if (this !== document.activeElement) {
                        searchInput.css({'opacity': 0});
                        searchImg.css({'opacity': 1});
                    }
                },
                focus: function () {
        			searchInput.css({'opacity': 1});
        			searchImg.css({'opacity': 0});
        		},
                blur: function () {
                    if (!mouseentered) {
            			searchInput.css({'opacity': 0});
            			searchImg.css({'opacity': 1});
                    }
        		}
            });
        },
        start: function () {
            spotify.getStatus(_.updatePlayingUi);

    		setInterval(function () {
    			spotify.getStatus(_.updatePlayingUi);
    		}, 1000);

    		spotify.refreshQueue();

    		setInterval(spotify.refreshQueue, 2000);
        }
    };

    return spotify;
})();
