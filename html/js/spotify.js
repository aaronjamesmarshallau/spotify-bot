/*jshint laxbreak: true */

var Spotify = (function () {
    var _ = {
        consts: {
            SEARCH_ENDPOINT: "/search?q=",
            ALBUMS_ENDPOINT: "/albums?id=",
            TRACKS_ENDPOINT: "/tracks?id="
        },
        currentStatus: null,
        currentQueue: [],
        currentClients: [],
        hasClientsChanged: function (newClients) {
            if (!newClients) {
                return false;
            }

            if (newClients.length !== _.currentClients.length) {
                return true;
            }

            var clientSort = function (a, b) {
                if (a.clientName < b.clientName) {
                    return -1;
                }
                if (a.clientName > b.clientName) {
                    return 1;
                }
                return 0;
            };

            newClients.sort(clientSort);
            _.currentClients.sort(clientSort);

            for (var i = 0; i < _.currentClients.length; i++) {
                var newClient = newClients[i];
                var oldClient = _.currentClients[i];

                if (newClient.clientToken !== oldClient.clientToken) {
                    return true;
                }
            }

            return false;
        },
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
            var loginButton = $("<button></button>", {
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
                        });
                    }
                }
            });

            return $("<div></div>", {
                class: "login-page",
                on: {
                    keypress: function (e) {
                        var keyCode = e.keyCode || e.which;
                        if (keyCode === 13) {
                            e.preventDefault();
                            loginButton.click();
                        }
                    }
                },
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
                                            loginButton,
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
        onTrackChanged: function (newTrack, resetVote) {
            if (!!resetVote) {
                window.localStorage.setItem("hasVoted", false);

                if (Notification) {
                    if (Notification.permission === "granted") {
                        var notification = new Notification(newTrack.trackName, {
                            icon: newTrack.albumArt.smallArt,
                            body: newTrack.artistName + "\n" + newTrack.albumName
                        });

                        notification.onshow = function () {
                            setTimeout(function () {
                                notification.close();
                            }, 2500);
                        };

                        notification.onclick = function () {
                            this.close();
                        };
                    }
                }
            }

			_.currentQueue.shift();
			_.onQueueChange(_.currentQueue);
			_.updateTrackUi();
        },
        onQueueChange: function (newQueue) {
            _.currentQueue = newQueue;
			$(".queue-list").empty();

			for (var i = 0; i < _.currentQueue.length; i++)
			(function (index) { //jshint ignore:line
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
                                            src: track.albumArt.smallArt,
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
        hashCode: function (str) {
            var hash = 0;
            for (var i = 0; i < str.length; i++) {
               hash = str.charCodeAt(i) + ((hash << 5) - hash);
            }
            return hash;
        },
        colorFromInt: function (i) {
            var c = (i & 0x00FFFFFF)
                .toString(16)
                .toUpperCase();

            return "00000".substring(0, 6 - c.length) + c;
        },
        updateIdentity: function () {
            $(".identity-text").text(window.localStorage.getItem("clientName"));
            $(".identity-image-letter").text(window.localStorage.getItem("clientName").replace("Anonymous ", "").substring(0, 1));
            $(".identity-image-inner").css("background-color", "#" + _.colorFromInt(_.hashCode(window.localStorage.getItem("clientId"))));
        },
        updatePlayingUi: function () {
            var playing = _.currentStatus.playing;
			var currentPlayPosition = (_.currentStatus.playPosition / _.currentStatus.nowPlaying.duration) * 100;

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
			var currentTrackTitle = _.currentStatus.nowPlaying.trackName;
			var currentArtistTitle = _.currentStatus.nowPlaying.artistName;
			var currentAlbumTitle = _.currentStatus.nowPlaying.albumName;
			var nowPlayingAreaHeight = $(".playing-panel-inner").height() - 20;
            var imageUri = _.currentStatus.nowPlaying.albumArt.largeArt;

			$(".album-artwork").attr("src", imageUri);
            $(".playing-panel-background").attr("src", imageUri);

			$(".now-playing-container").height(nowPlayingAreaHeight);
			$(".now-playing-container").width(nowPlayingAreaHeight);

			$(".song-title-inner").text(currentTrackTitle);
			$(".artist-title-inner").text(currentArtistTitle);
			$(".album-title-inner").text(currentAlbumTitle);
        },
        onClientsChanged: function (clients) {
            var maxNumOfClients = 3;
            var clientElements = [];
            var clientsToMakeElsFor = [];
            var i = 0;
            _.currentClients = clients;

            for (i = 0; i < clients.length; i++) {
                (function (ind) { /*jshint ignore: line */
                    var cl = clients[ind];
                    if (cl.identityToken !== window.localStorage.getItem("clientId")) {
                        clientsToMakeElsFor.push(cl);
                    }
                })(i);
            }

            if (clientsToMakeElsFor.length < maxNumOfClients) {
                maxNumOfClients = clientsToMakeElsFor.length;
            }

            for (i = 0; i < maxNumOfClients; i++)
            (function (index) { /*jshint ignore:line */
                var client = clientsToMakeElsFor[i];
                var clientName = client.identityName.replace("Anonymous ", "");
                var clientToken = client.identityToken;
                var clientColor = "#" + _.colorFromInt(_.hashCode(clientToken));

                var clientEl = $("<div></div>", {
                    class: "identity-image-container",
                    html: [
                        $("<div></div>", {
                            class: "identity-image-inner",
                            style: "background-color:" + clientColor,
                            "data-client-id": clientToken,
                            title: clientName,
                            html: [
                                $("<div></div>", {
                                    class: "identity-image-letter",
                                    text: clientName.substring(0, 1)
                                })
                            ]
                        })
                    ]
                });

                clientElements.push(clientEl);
            })(i);

            if (clientsToMakeElsFor.length > maxNumOfClients) {
                var difference = clientsToMakeElsFor.length - maxNumOfClients;
                var text = "+" + difference;
                var clientColor = "#FFBB00";
                var clientEl = $("<div></div>", {
                    class: "identity-image-container",
                    html: [
                        $("<div></div>", {
                            class: "identity-image-inner",
                            style: "background-color:" + clientColor,
                            title: text + ' more',
                            html: [
                                $("<div></div>", {
                                    class: "identity-image-letter",
                                    text: text
                                })
                            ]
                        })
                    ]
                });

                clientElements.splice(0, 0, clientEl);
            }

            $(".connected-client-container").empty();
            $(".connected-client-container").append(clientElements);
        }
    };

    var spotify = {
        ajax: function (config) {
            var me = this;
            var clientSecret = window.localStorage.getItem("clientSecret");
            var clientId = window.localStorage.getItem("clientId");
            var callback = function () {};

            if (config.success != null) { /*jshint ignore: line */
                callback = config.success;
            }

            config.success = function (response) {
                if (response != null && response.responseStatus === 3) {
                    me.registerClient(function () {
                        me.ajax(config);
                    });
                    return;
                }

                if (typeof callback === "function") {
                    callback(response);
                }
            };

            if (!!clientId && !!clientSecret) {
                if (!config.headers) {
                    config.headers = {};
                }

                config.headers["X-Client-Token"] = clientId;
                config.headers["X-Client-Secret"] = clientSecret;
            }

            $.ajax(config);
        },
        registerClient: function (callback) {
            var me = this;
            var clientSecret = window.localStorage.getItem("clientSecret");
            var headers = {};

            me.ajax({
                url: "/identify",
                method: "GET",
                success: function (response) {
                    var identityToken = response.identityToken;
                    var identityName = response.identityName;
                    var identitySecret = response.identitySecret;

                    window.localStorage.setItem("clientId", identityToken);
                    window.localStorage.setItem("clientName", identityName);
                    window.localStorage.setItem("clientSecret", identitySecret);

                    if (typeof callback === "function") callback();
                }
            });
        },
        getClients: function (callback) {
            var me = this;

            me.ajax({
                url: "/clients",
                success: function (response) {
                    if (_.hasClientsChanged(response)) {
                        callback(response);
                    }
                }
            });
        },
        getStatus: function (callback) {
            var me = this;

            me.ajax({
				url: "/status",
				method: "GET",
				success: function (data) {
                    var newStatus = data;
                    var newTrack = newStatus.nowPlaying;
                    var newTrackId = newTrack.trackId;

					if (_.currentStatus != null) { /*jshint ignore: line */
                        var currentTrack = _.currentStatus.nowPlaying.trackId;

                        // Old status was further along in the song than the new song, and the new playing position is
                        // within the first five seconds of playback.
                        var trackHasReset = _.currentStatus.playPosition > newStatus.playPosition && newStatus.playPosition < 5 && newStatus.playing;

                        _.currentStatus = data;

						if (currentTrack !== newTrackId || trackHasReset) {
							_.onTrackChanged(newTrack, true);
						}
					} else {
						_.currentStatus = data;
						_.onTrackChanged(newTrack, false);
					}

					if (typeof callback === "function") callback();
				}
			});
        },
        getTrackInfo: function (trackId, callback) {
            me.ajax({
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
            spotify.ajax({
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
            var me = this;

            if (text.length === 0) {
				$(".search-results").remove();
				return;
			}

			me.ajax({
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
						(function (index) { //jshint ignore:line
							var el = tracks[index];
							var trackId = el.id;
                            var albumName = el.album.name;
                            var duration = (el.duration_ms / 1000);
                            var albumArtColl = {
                                smallArt: el.album.images[2].url,
                                mediumArt: el.album.images[1].url,
                                largeArt: el.album.images[0].url,
                            };
							var albumArtSmall = albumArtColl.smallArt;
                            var albumArt = albumArtColl.largeArt;
							var trackName = el.name;
							var trackArtist = el.artists[0].name;
							var queueUrl = "/queue";

							var resultEl = $("<li></li>", {
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
                                                        var trackInfo = {
                                                            trackId: trackId,
                                                            trackName: trackName,
                                                            artistName: trackArtist,
                                                            albumArt: albumArtColl,
                                                            albumName: albumName,
                                                            duration: duration
                                                        };

														me.ajax({
															url: queueUrl,
                                                            method: "POST",
                                                            dataType: "json",
                                                            contentType: "application/json; charset=UTF-8",
                                                            data: JSON.stringify(trackInfo),
															success: function (data) {
                                                                if (data.success) {
            														_.onQueueChange(data);

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
    																});
                                                                } else {
                                                                    if (data.message) {
                                                                        alert(data.message);
                                                                    } else {
                                                                        alert("Something went wrong...");
                                                                    }
                                                                }
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

							results.push(resultEl);
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
        attemptRegister: function (authCallback) {
            var me = this;

            me.ajax({
        		url: "/auth",
        		method: "GET",
        		success: function (response, textStatus) {
        			if (response.success && typeof authCallback === "function") {
                        authCallback();
        			}
        		}
        	});
        },
        init: function () {
            $(".playpause").on('click', function () {
    			var endpoint = _.currentStatus != null /*jshint ignore: line */
    				? _.currentStatus.playing
    					? "/pause"
    					: "/unpause"
    				: "/unpause";
    			spotify.ajax({
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

    			spotify.ajax({
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

    			spotify.ajax({
    				url: "/upvote"
    			});
    		});


    		var searchInput = $(".spotify-search input");
    		var searchImg = $(".spotify-search img");
            var timeout;
            var mouseentered = false;

            searchInput.on({
                keyup: function (evt) {
                    if (evt.keyCode === 27) {
                        $(this).val("");
                        $('.search-results').remove();
                        return;
                    }

        			var text = $(this).val();

                    if (timeout != null) { /*jshint ignore: line */
                        clearTimeout(timeout);
                    }

        			timeout = setTimeout(function () {
        				spotify.search(text);
        			}, 250);
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

            // If we have notification capabilities
            if (!!Notification) {
                if (Notification.permission !== "granted" && !localStorage.getItem("notificationRequested")) {
                    localStorage.setItem("notificationRequested", "true");

                    Notification.requestPermission();
                }
            }
        },
        start: function () {
            spotify.registerClient(_.updateIdentity);
            spotify.getStatus(_.updatePlayingUi);
            spotify.getClients(_.onClientsChanged);
            spotify.attemptRegister();

    		setInterval(function () {
    			spotify.getStatus(_.updatePlayingUi);
    		}, 1000);

            setInterval(function () {
                spotify.getClients(_.onClientsChanged);
            }, 5000);

    		spotify.refreshQueue();

    		setInterval(spotify.refreshQueue, 2000);
        }
    };

    return spotify;
})();
