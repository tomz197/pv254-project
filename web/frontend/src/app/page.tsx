import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { CourseSearch } from "@/components/course-search";
import { SelectedCourses } from "@/components/selected-courses";
import type {
  CoursePreferences,
  CourseSearch as CourseSearchType,
} from "@/types";
import { useNavigate } from "react-router";
import { storageController } from "@/storage";

export default function Home() {
  const navigate = useNavigate();
  const [likedCourses, setLikedCourses] = useState<CoursePreferences["liked"]>(
    new Map(),
  );
  const [dislikedCourses, setDislikedCourses] = useState<
    CoursePreferences["disliked"]
  >(new Map());

  useEffect(() => {
    const { liked, disliked } = storageController.getCoursePreferences();

    setLikedCourses(liked);
    setDislikedCourses(disliked);
  }, []);

  const handleAddLikedCourse = (course: CourseSearchType) => {
    if (likedCourses.has(course.CODE)) return;
    setLikedCourses(new Map(likedCourses.set(course.CODE, course)));
  };

  const handleAddDislikedCourse = (course: CourseSearchType) => {
    if (dislikedCourses.has(course.CODE)) return;
    setDislikedCourses(new Map(dislikedCourses.set(course.CODE, course)));
  };

  const handleRemoveLikedCourse = (courseId: string) => {
    likedCourses.delete(courseId);
    setLikedCourses(new Map(likedCourses));
  };

  const handleRemoveDislikedCourse = (courseId: string) => {
    dislikedCourses.delete(courseId);
    setDislikedCourses(new Map(dislikedCourses));
  };

  const handleGetRecommendations = () => {
    storageController.setCoursePreferences({
      liked: likedCourses,
      disliked: dislikedCourses,
    });
    navigate("/recommendations");
  };

  return (
    <main className="container mx-auto px-4 py-8 md:py-12">
      <div className="flex flex-col items-center text-center mb-12">
        <h1 className="text-3xl md:text-4xl lg:text-5xl font-bold tracking-tight mb-4">
          MUNI Course Recommendation
        </h1>
        <p className="text-lg text-muted-foreground max-w-2xl">
          Help us understand your preferences by selecting courses you like and
          dislike, and we'll recommend courses tailored to your interests.
        </p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-8 mb-12">
        <Card className="h-full">
          <CardHeader>
            <CardTitle>Courses You Like</CardTitle>
            <CardDescription>
              Select courses that interest you or that you've enjoyed in the
              past
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4 flex flex-col">
            <div className="flex-1">
              <SelectedCourses
                courses={Array.from(likedCourses.values())}
                onRemove={handleRemoveLikedCourse}
                emptyMessage="No liked courses selected yet"
              />
            </div>
            <CourseSearch
              onSelectCourse={handleAddLikedCourse}
              placeholder="Search for courses you like..."
              excludeCourses={[
                ...likedCourses.values(),
                ...dislikedCourses.values(),
              ]}
            />
          </CardContent>
        </Card>

        <Card className="h-full">
          <CardHeader>
            <CardTitle>Courses You Dislike</CardTitle>
            <CardDescription>
              Select courses that don't interest you or that you didn't enjoy
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4 flex flex-col">
            <div className="flex-1">
              <SelectedCourses
                courses={Array.from(dislikedCourses.values())}
                onRemove={handleRemoveDislikedCourse}
                emptyMessage="No disliked courses selected yet"
              />
            </div>
            <CourseSearch
              onSelectCourse={handleAddDislikedCourse}
              placeholder="Search for courses you dislike..."
              excludeCourses={[
                ...likedCourses.values(),
                ...dislikedCourses.values(),
              ]}
            />
          </CardContent>
        </Card>
      </div>

      <div className="flex justify-center">
        <Button
          size="lg"
          onClick={handleGetRecommendations}
          disabled={likedCourses.size === 0}
        >
          Get Course Recommendations
        </Button>
      </div>
    </main>
  );
}
